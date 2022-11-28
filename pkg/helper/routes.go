package helper

import (
	"github.com/authgear/authgear-delete-user-helper/pkg/helper/deps"
	"github.com/authgear/authgear-delete-user-helper/pkg/helper/handler"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func NewRouter(p *deps.RootProvider, configSource *configsource.ConfigSource, auth config.AdminAPIAuth) *httproute.Router {
	router := httproute.NewRouter()

	router.Add(httproute.Route{
		Methods:     []string{"GET"},
		PathPattern: "/healthz",
	}, p.RootHandler(newHealthzHandler))

	securityMiddleware := httproute.Chain(
		httproute.MiddlewareFunc(httputil.StaticSecurityHeaders),
		httputil.StaticCSPHeader{
			CSPDirectives: []string{
				"script-src 'self' 'unsafe-inline' unpkg.com",
				"object-src 'none'",
				"base-uri 'none'",
				"block-all-mixed-content",
				"frame-ancestors 'none'",
			},
		},
	)

	chain := httproute.Chain(
		p.RootMiddleware(newPanicMiddleware),
		p.RootMiddleware(newBodyLimitMiddleware),
		p.RootMiddleware(newSentryMiddleware),
		securityMiddleware,
		httproute.MiddlewareFunc(httputil.NoStore),
		&deps.RequestMiddleware{
			RootProvider: p,
			ConfigSource: configSource,
		},
		p.Middleware(func(p *deps.RequestProvider) httproute.Middleware {
			return newAuthorizationMiddleware(p, auth)
		}),
	)

	route := httproute.Route{Middleware: chain}

	router.AddRoutes(p.Handler(newSearchUserHandler), handler.ConfigureSearchUserRoute(route))

	return router
}
