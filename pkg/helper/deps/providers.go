package deps

import (
	"context"
	"net/http"

	getsentry "github.com/getsentry/sentry-go"
	"github.com/spf13/afero"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/resource"
	"github.com/authgear/authgear-server/pkg/util/sentry"
)

func NewManager(registry *resource.Registry, customResourceDir string) *resource.Manager {
	var fs []resource.Fs
	if customResourceDir != "" {
		fs = append(fs,
			resource.LeveledAferoFs{
				Fs:      afero.NewBasePathFs(afero.OsFs{}, customResourceDir),
				FsLevel: resource.FsLevelCustom,
			},
		)
	}
	return &resource.Manager{
		Registry: registry.Clone(),
		Fs:       fs,
	}
}

type RootProvider struct {
	EnvironmentConfig  *config.EnvironmentConfig
	ConfigSourceConfig *configsource.Config
	LoggerFactory      *log.Factory
	SentryHub          *getsentry.Hub
	DatabasePool       *db.Pool
	RedisPool          *redis.Pool
	BaseResources      *resource.Manager
}

func NewRootProvider(
	cfg *config.EnvironmentConfig,
	configSourceConfig *configsource.Config,
	customResourceDirectory string,
) (*RootProvider, error) {
	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	sentryHub, err := sentry.NewHub(string(cfg.SentryDSN))
	if err != nil {
		return nil, err
	}

	loggerFactory := log.NewFactory(
		logLevel,
		log.NewDefaultMaskLogHook(),
		sentry.NewLogHookFromHub(sentryHub),
	)

	dbPool := db.NewPool()
	redisPool := redis.NewPool()

	resourceManager := NewManager(
		resource.DefaultRegistry,
		customResourceDirectory,
	)

	return &RootProvider{
		EnvironmentConfig:  cfg,
		ConfigSourceConfig: configSourceConfig,
		LoggerFactory:      loggerFactory,
		SentryHub:          sentryHub,
		DatabasePool:       dbPool,
		RedisPool:          redisPool,
		BaseResources:      resourceManager,
	}, nil
}

func (p *RootProvider) NewAppProvider(ctx context.Context, appCtx *config.AppContext) *AppProvider {
	cfg := appCtx.Config
	loggerFactory := p.LoggerFactory.ReplaceHooks(
		log.NewDefaultMaskLogHook(),
		config.NewSecretMaskLogHook(cfg.SecretConfig),
		sentry.NewLogHookFromContext(ctx),
	)
	loggerFactory.DefaultFields["app"] = cfg.AppConfig.ID
	appDatabase := appdb.NewHandle(
		ctx,
		p.DatabasePool,
		&p.EnvironmentConfig.DatabaseConfig,
		cfg.SecretConfig.LookupData(config.DatabaseCredentialsKey).(*config.DatabaseCredentials),
		loggerFactory,
	)
	provider := &AppProvider{
		RootProvider:  p,
		Context:       ctx,
		Config:        cfg,
		LoggerFactory: loggerFactory,
		AppDatabase:   appDatabase,
	}
	return provider
}

func (p *RootProvider) RootHandler(factory func(*RootProvider, http.ResponseWriter, *http.Request, context.Context) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := factory(p, w, r, r.Context())
		h.ServeHTTP(w, r)
	})
}

func (p *RootProvider) RootMiddleware(factory func(*RootProvider) httproute.Middleware) httproute.Middleware {
	return factory(p)
}

func (p *RootProvider) Middleware(f func(*RequestProvider) httproute.Middleware) httproute.Middleware {
	return httproute.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestProvider := getRequestProvider(w, r)
			m := f(requestProvider)
			h := m.Handle(next)
			h.ServeHTTP(w, r)
		})
	})
}

func (p *RootProvider) Handler(f func(*RequestProvider) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestProvider := getRequestProvider(w, r)
		h := f(requestProvider)
		h.ServeHTTP(w, r)
	})
}

type AppProvider struct {
	*RootProvider
	Context       context.Context
	Config        *config.Config
	LoggerFactory *log.Factory
	AppDatabase   *appdb.Handle
}

func (p *AppProvider) NewRequestProvider(r *http.Request) *RequestProvider {
	return &RequestProvider{
		AppProvider: p,
		Request:     r,
	}
}

type RequestProvider struct {
	*AppProvider
	Request *http.Request
}
