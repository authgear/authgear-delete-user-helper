package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

var envConfigDeps = wire.NewSet(
	wire.FieldsOf(new(*config.EnvironmentConfig),
		"TrustProxy",
		"DevMode",
		"SentryDSN",
		"GlobalDatabase",
		"DatabaseConfig",
		"ImagesCDNHost",
		"WebAppCDNHost",
		"CORSAllowedOrigins",
		"RedisConfig",
		"NFTIndexerAPIEndpoint",
	),
)

var RootDependencySet = wire.NewSet(
	wire.FieldsOf(new(*RootProvider),
		"EnvironmentConfig",
		"ConfigSourceConfig",
		"LoggerFactory",
		"SentryHub",
		"DatabasePool",
		"RedisPool",
		"BaseResources",
	),
	envConfigDeps,

	clock.DependencySet,
	globaldb.DependencySet,
	configsource.DependencySet,
)

var AppRootDependencySet = wire.NewSet(
	RootDependencySet,
	wire.FieldsOf(new(*AppProvider),
		"RootProvider",
		"Config",
		"AppDatabase",
	),
)

var RequestDependencySet = wire.NewSet(
	AppRootDependencySet,
	wire.FieldsOf(new(*RequestProvider),
		"AppProvider",
		"Request",
	),
	deps.ProvideRequestContext,
	deps.ProvideRemoteIP,
	deps.ProvideUserAgentString,
	deps.ProvideHTTPHost,
	deps.ProvideHTTPProto,
)

var DependencySet = wire.NewSet(
	RequestDependencySet,
	deps.CommonDependencySet,
)
