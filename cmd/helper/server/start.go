package server

import (
	"context"
	golog "log"

	helper "github.com/authgear/authgear-delete-user-helper/pkg/helper"
	helperdeps "github.com/authgear/authgear-delete-user-helper/pkg/helper/deps"
	"github.com/authgear/authgear-server/pkg/util/server"
)

func Start() error {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		golog.Fatalf("failed to load server config: %v", err)
	}

	p, err := helperdeps.NewRootProvider(
		cfg.EnvironmentConfig,
		cfg.ConfigSource,
	)
	if err != nil {
		golog.Fatalf("failed to setup server: %v", err)
	}

	logger := p.LoggerFactory.New("server")

	configSrcController := newConfigSourceController(p, context.Background())
	err = configSrcController.Open()
	if err != nil {
		logger.WithError(err).Fatal("cannot open configuration")
	}
	defer configSrcController.Close()

	u, err := server.ParseListenAddress(cfg.DeleteUserHelperListenAddr)
	if err != nil {
		logger.WithError(err).Fatal("failed to parse admin API server listen address")
	}

	specs := []server.Spec{
		{
			Name:          "Delete User Helper",
			ListenAddress: u.Host,
			Handler: helper.NewRouter(
				p,
				configSrcController.GetConfigSource(),
				cfg.AdminAPIAuth,
			),
		},
	}

	server.Start(logger, specs)
	return nil
}
