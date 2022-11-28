package helper

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-delete-user-helper/pkg/helper/deps"
	"github.com/authgear/authgear-delete-user-helper/pkg/helper/handler"
	adminauthz "github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/infra/middleware"
)

var DependencySet = wire.NewSet(
	deps.DependencySet,
	middleware.DependencySet,
	adminauthz.DependencySet,
	handler.DependencySet,
)
