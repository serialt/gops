package repo

import (
	"context"

	"gops/internal/data/define"
	"gops/internal/data/model"
	"gops/internal/data/query"
	"gops/internal/pkg/utils"

	"gorm.io/gen/field"
	"gorm.io/gorm"
)

type AssetRepo struct{}

func NewAssetRepo() *AssetRepo { return &AssetRepo{} }

func (r *AssetRepo) ListGroups(ctx context.Context, db *gorm.DB, cond *define.AssetGroupListCond) ([]*model.AssetGroup, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).AssetGroup
	do := q.WithContext(ctx)
	if cond.Keyword != "" {
		value := "%" + cond.Keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.Code.Like(value)))
	}
	if cond.ParentID > 0 {
		do = do.Where(q.ParentID.Eq(cond.ParentID))
	}
	return do.Order(q.SortOrder, q.ID).FindByPage(offset, limit)
}

func (r *AssetRepo) ListCredentials(ctx context.Context, db *gorm.DB, cond *define.CredentialListCond) ([]*model.Credential, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).Credential
	do := q.WithContext(ctx)
	if cond.Keyword != "" {
		value := "%" + cond.Keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.Username.Like(value)))
	}
	return do.Order(q.ID.Desc()).FindByPage(offset, limit)
}

func (r *AssetRepo) ListCloudAccounts(ctx context.Context, db *gorm.DB, cond *define.CloudAccountListCond) ([]*model.CloudAccount, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).CloudAccount
	do := q.WithContext(ctx)
	if cond.Keyword != "" {
		value := "%" + cond.Keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.Provider.Like(value)))
	}
	return do.Order(q.ID.Desc()).FindByPage(offset, limit)
}

func (r *AssetRepo) ListHosts(ctx context.Context, db *gorm.DB, cond *define.HostListCond) ([]*model.Host, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).Host
	do := q.WithContext(ctx)
	if cond.Keyword != "" {
		value := "%" + cond.Keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.IP.Like(value)))
	}
	if cond.GroupID > 0 {
		do = do.Where(q.GroupID.Eq(cond.GroupID))
	}
	if cond.Status != nil {
		do = do.Where(q.Status.Eq(*cond.Status))
	}
	if len(cond.IDs) > 0 {
		do = do.Where(q.ID.In(cond.IDs...))
	}
	return do.Order(q.ID.Desc()).FindByPage(offset, limit)
}

func (r *AssetRepo) GetHostByIP(ctx context.Context, db *gorm.DB, ip string) (*model.Host, error) {
	q := query.Use(db).Host
	return q.WithContext(ctx).Where(q.IP.Eq(ip)).First()
}

func (r *AssetRepo) GetHostAndVNCCredential(ctx context.Context, db *gorm.DB, id int64) (*model.Host, *model.Credential, error) {
	q := query.Use(db)
	host, err := q.Host.WithContext(ctx).Where(q.Host.ID.Eq(id)).First()
	if err != nil {
		return nil, nil, err
	}
	if host.VncCredentialID == nil || *host.VncCredentialID <= 0 {
		return nil, nil, gorm.ErrRecordNotFound
	}
	credential, err := q.Credential.WithContext(ctx).Where(q.Credential.ID.Eq(*host.VncCredentialID)).First()
	if err != nil {
		return nil, nil, err
	}
	return host, credential, nil
}

func (r *AssetRepo) CreateHost(ctx context.Context, db *gorm.DB, host *model.Host) error {
	return query.Use(db).Host.WithContext(ctx).Create(host)
}

func (r *AssetRepo) SaveHost(ctx context.Context, db *gorm.DB, host *model.Host) error {
	q := query.Use(db).Host
	return q.WithContext(ctx).Where(q.ID.Eq(host.ID)).Save(host)
}

func (r *AssetRepo) ListDBInstances(ctx context.Context, db *gorm.DB, cond *define.DBInstanceListCond) ([]*model.DbInstance, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).DbInstance
	do := q.WithContext(ctx)
	if cond.Keyword != "" {
		value := "%" + cond.Keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.Host.Like(value)))
	}
	if cond.GroupID > 0 {
		do = do.Where(q.GroupID.Eq(cond.GroupID))
	}
	if cond.DbType != "" {
		do = do.Where(q.DbType.Eq(cond.DbType))
	}
	if len(cond.IDs) > 0 {
		do = do.Where(q.ID.In(cond.IDs...))
	}
	return do.Order(q.ID.Desc()).FindByPage(offset, limit)
}

func (r *AssetRepo) GetDBInstanceByEndpoint(ctx context.Context, db *gorm.DB, dbType, host string, port int64) (*model.DbInstance, error) {
	q := query.Use(db).DbInstance
	return q.WithContext(ctx).Where(q.DbType.Eq(dbType), q.Host.Eq(host), q.Port.Eq(port)).First()
}

func (r *AssetRepo) CreateDBInstance(ctx context.Context, db *gorm.DB, instance *model.DbInstance) error {
	return query.Use(db).DbInstance.WithContext(ctx).Create(instance)
}

func (r *AssetRepo) SaveDBInstance(ctx context.Context, db *gorm.DB, instance *model.DbInstance) error {
	q := query.Use(db).DbInstance
	return q.WithContext(ctx).Where(q.ID.Eq(instance.ID)).Save(instance)
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
