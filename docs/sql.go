package docs

import "embed"

//go:embed sql/*.sql
var Migrations embed.FS
