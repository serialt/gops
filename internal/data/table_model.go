package data

import (
	"gorm.io/gen"
	"gorm.io/gen/field"
)

func gopsTable(g *gen.Generator) {
	role := g.GenerateModel("role")
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
	))
	g.ApplyBasic(
		role,
		userRole,
		user,
	)

}
