package data

import (
	"gorm.io/gen"
	"gorm.io/gen/field"
)

func gopsTable(g *gen.Generator) {
	role := g.GenerateModel("role")
	organization := g.GenerateModel("organization")
	organizationUser := g.GenerateModel("organization_user")
	systemConfig := g.GenerateModel("system_config")
	assetGroup := g.GenerateModel("asset_group")
	credential := g.GenerateModel("credential")
	cloudAccount := g.GenerateModel("cloud_account")
	host := g.GenerateModel("host")
	dbInstance := g.GenerateModel("db_instance")
	terminalSession := g.GenerateModel("terminal_session")
	userRole := g.GenerateModel("user_role")
	user := g.GenerateModel("user", gen.FieldRelate(field.HasMany, "Roles", role,
		&field.RelateConfig{
			RelateSlicePointer: true,
			GORMTag: field.GormTag{
				"many2many": []string{
					"user_role",
				},
			},
		},
	), gen.FieldRelate(field.HasMany, "Organizations", organization,
		&field.RelateConfig{RelateSlicePointer: true, GORMTag: field.GormTag{"many2many": []string{"organization_user"}}},
	))
	g.ApplyBasic(
		role,
		userRole,
		user,
		organization,
		organizationUser,
		systemConfig,
		assetGroup,
		credential,
		cloudAccount,
		host,
		dbInstance,
		terminalSession,
	)

}
