package repo

import (
	"context"

	"gops/internal/data/model"
	"gops/internal/data/query"

	"gorm.io/gorm"
)

type AssetRepo struct{}

func NewAssetRepo() *AssetRepo { return &AssetRepo{} }

func (r *AssetRepo) GetCredential(ctx context.Context, db *gorm.DB, id int64) (*model.Credential, error) {
	q := query.Use(db)
	return q.Credential.WithContext(ctx).Where(q.Credential.ID.Eq(id)).First()
}

func (r *AssetRepo) GetDBInstanceAndCredential(ctx context.Context, db *gorm.DB, id int64) (*model.DbInstance, *model.Credential, error) {
	q := query.Use(db)
	instance, err := q.DbInstance.WithContext(ctx).Where(q.DbInstance.ID.Eq(id)).First()
	if err != nil {
		return nil, nil, err
	}
	credential, err := q.Credential.WithContext(ctx).Where(q.Credential.ID.Eq(instance.CredentialID)).First()
	if err != nil {
		return nil, nil, err
	}
	return instance, credential, nil
}

func (r *AssetRepo) UpdateDBInstanceCollectedInfo(ctx context.Context, db *gorm.DB, id int64, status bool, version string) (*model.DbInstance, error) {
	q := query.Use(db)
	updates := map[string]any{"status": status}
	if version != "" {
		updates["version"] = version
	}
	if _, err := q.DbInstance.WithContext(ctx).Where(q.DbInstance.ID.Eq(id)).Updates(updates); err != nil {
		return nil, err
	}
	return q.DbInstance.WithContext(ctx).Where(q.DbInstance.ID.Eq(id)).First()
}
