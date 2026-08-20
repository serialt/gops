package repo

import (
	"gops/internal/data/repo/rbac"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	rbac.NewRbacRepo,
)
