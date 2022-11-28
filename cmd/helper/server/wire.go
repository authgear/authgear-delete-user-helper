//go:build wireinject
// +build wireinject

package server

import (
	"context"
	"github.com/google/wire"

	"github.com/authgear/authgear-delete-user-helper/pkg/helper/deps"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
)

func newConfigSourceController(p *deps.RootProvider, c context.Context) *configsource.Controller {
	panic(wire.Build(
		configsource.NewResolveAppIDTypeDomain,
		deps.RootDependencySet,
		configsource.ControllerDependencySet,
	))
}
