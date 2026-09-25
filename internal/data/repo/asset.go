package repo

import (
	"context"

	"gops/internal/data/model"
	"gops/internal/data/query"

	"gorm.io/gen/field"
	"gorm.io/gorm"
)

type AssetRepo struct{}

func NewAssetRepo() *AssetRepo { return &AssetRepo{} }

func (r *AssetRepo) ListGroups(ctx context.Context, db *gorm.DB, keyword string, offset, limit int) ([]*model.AssetGroup, int64, error) {
	q := query.Use(db).AssetGroup
	do := q.WithContext(ctx)
	if keyword != "" {
		value := "%" + keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.Code.Like(value)))
	}
	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}
	rows, err := do.Order(q.SortOrder, q.ID).Offset(offset).Limit(limit).Find()
	return rows, total, err
}

func (r *AssetRepo) ListCredentials(ctx context.Context, db *gorm.DB, keyword string, offset, limit int) ([]*model.Credential, int64, error) {
	q := query.Use(db).Credential
	do := q.WithContext(ctx)
	if keyword != "" {
		value := "%" + keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.Username.Like(value)))
	}
	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}
	rows, err := do.Order(q.ID.Desc()).Offset(offset).Limit(limit).Find()
	return rows, total, err
}

func (r *AssetRepo) ListCloudAccounts(ctx context.Context, db *gorm.DB, keyword string, offset, limit int) ([]*model.CloudAccount, int64, error) {
	q := query.Use(db).CloudAccount
	do := q.WithContext(ctx)
	if keyword != "" {
		value := "%" + keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.Provider.Like(value)))
	}
	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}
	rows, err := do.Order(q.ID.Desc()).Offset(offset).Limit(limit).Find()
	return rows, total, err
}

func (r *AssetRepo) ListHosts(ctx context.Context, db *gorm.DB, keyword string, groupID int64, offset, limit int) ([]*model.Host, int64, error) {
	q := query.Use(db).Host
	do := q.WithContext(ctx)
	if keyword != "" {
		value := "%" + keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.IP.Like(value)))
	}
	if groupID > 0 {
		do = do.Where(q.GroupID.Eq(groupID))
	}
	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}
	rows, err := do.Order(q.ID.Desc()).Offset(offset).Limit(limit).Find()
	return rows, total, err
}

func (r *AssetRepo) ListDBInstances(ctx context.Context, db *gorm.DB, keyword string, groupID int64, offset, limit int) ([]*model.DbInstance, int64, error) {
	q := query.Use(db).DbInstance
	do := q.WithContext(ctx)
	if keyword != "" {
		value := "%" + keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.Host.Like(value)))
	}
	if groupID > 0 {
		do = do.Where(q.GroupID.Eq(groupID))
	}
	total, err := do.Count()
	if err != nil {
		return nil, 0, err
	}
	rows, err := do.Order(q.ID.Desc()).Offset(offset).Limit(limit).Find()
	return rows, total, err
}

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
