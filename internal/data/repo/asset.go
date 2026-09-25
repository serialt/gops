package repo

import (
	"context"

	"gops/internal/data/model"
	"gops/internal/data/query"

	"gorm.io/gorm"
)

type AssetRepo struct{ db *gorm.DB }

func NewAssetRepo(db *gorm.DB) *AssetRepo { return &AssetRepo{db: db} }

func (r *AssetRepo) GetCredential(ctx context.Context, id int64) (*model.Credential, error) {
	q := query.Use(r.db)
	return q.Credential.WithContext(ctx).Where(q.Credential.ID.Eq(id)).First()
}

func (r *AssetRepo) GetDBInstanceAndCredential(ctx context.Context, id int64) (*model.DbInstance, *model.Credential, error) {
	q := query.Use(r.db)
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

func (r *AssetRepo) UpdateDBInstanceCollectedInfo(ctx context.Context, id int64, status bool, version string) (*model.DbInstance, error) {
	q := query.Use(r.db)
	updates := map[string]any{"status": status}
	if version != "" {
		updates["version"] = version
	}
	if _, err := q.DbInstance.WithContext(ctx).Where(q.DbInstance.ID.Eq(id)).Updates(updates); err != nil {
		return nil, err
	}
	return q.DbInstance.WithContext(ctx).Where(q.DbInstance.ID.Eq(id)).First()
}
