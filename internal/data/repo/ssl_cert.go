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

type SSLCertRepo struct{}

func NewSSLCertRepo() *SSLCertRepo { return &SSLCertRepo{} }

func (r *SSLCertRepo) ListCertificates(ctx context.Context, db *gorm.DB, cond *define.SSLCertificateListCond) ([]*model.SslCertificate, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).SslCertificate
	do := q.WithContext(ctx)
	if cond.Keyword != "" {
		value := "%" + cond.Keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.Domain.Like(value)))
	}
	if cond.Status != "" {
		do = do.Where(q.Status.Eq(cond.Status))
	}
	if cond.SourceType != "" {
		do = do.Where(q.SourceType.Eq(cond.SourceType))
	}
	return do.Order(q.ID.Desc()).FindByPage(offset, limit)
}

func (r *SSLCertRepo) GetCertificate(ctx context.Context, db *gorm.DB, id int64) (*model.SslCertificate, error) {
	q := query.Use(db).SslCertificate
	return q.WithContext(ctx).Where(q.ID.Eq(id)).First()
}

func (r *SSLCertRepo) CreateCertificate(ctx context.Context, db *gorm.DB, row *model.SslCertificate) error {
	return query.Use(db).SslCertificate.WithContext(ctx).Create(row)
}

func (r *SSLCertRepo) SaveCertificate(ctx context.Context, db *gorm.DB, row *model.SslCertificate) error {
	q := query.Use(db).SslCertificate
	return q.WithContext(ctx).Where(q.ID.Eq(row.ID)).Save(row)
}

func (r *SSLCertRepo) DeleteCertificate(ctx context.Context, db *gorm.DB, row *model.SslCertificate) error {
	_, err := query.Use(db).SslCertificate.WithContext(ctx).Delete(row)
	return err
}

func (r *SSLCertRepo) ListDNSProviders(ctx context.Context, db *gorm.DB, cond *define.SSLDNSProviderListCond) ([]*model.SslDNSProvider, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).SslDNSProvider
	do := q.WithContext(ctx)
	if cond.Keyword != "" {
		do = do.Where(q.Name.Like("%" + cond.Keyword + "%"))
	}
	if cond.Provider != "" {
		do = do.Where(q.Provider.Eq(cond.Provider))
	}
	return do.Order(q.ID.Desc()).FindByPage(offset, limit)
}

func (r *SSLCertRepo) GetDNSProvider(ctx context.Context, db *gorm.DB, id int64) (*model.SslDNSProvider, error) {
	q := query.Use(db).SslDNSProvider
	return q.WithContext(ctx).Where(q.ID.Eq(id)).First()
}
func (r *SSLCertRepo) CreateDNSProvider(ctx context.Context, db *gorm.DB, row *model.SslDNSProvider) error {
	return query.Use(db).SslDNSProvider.WithContext(ctx).Create(row)
}
func (r *SSLCertRepo) SaveDNSProvider(ctx context.Context, db *gorm.DB, row *model.SslDNSProvider) error {
	q := query.Use(db).SslDNSProvider
	return q.WithContext(ctx).Where(q.ID.Eq(row.ID)).Save(row)
}
func (r *SSLCertRepo) DeleteDNSProvider(ctx context.Context, db *gorm.DB, row *model.SslDNSProvider) error {
	_, err := query.Use(db).SslDNSProvider.WithContext(ctx).Delete(row)
	return err
}

func (r *SSLCertRepo) ListDeployConfigs(ctx context.Context, db *gorm.DB, cond *define.SSLDeployConfigListCond) ([]*model.SslDeployConfig, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).SslDeployConfig
	do := q.WithContext(ctx)
	if cond.Keyword != "" {
		do = do.Where(q.Name.Like("%" + cond.Keyword + "%"))
	}
	if cond.CertificateID > 0 {
		do = do.Where(q.CertificateID.Eq(cond.CertificateID))
	}
	if cond.DeployType != "" {
		do = do.Where(q.DeployType.Eq(cond.DeployType))
	}
	return do.Order(q.ID.Desc()).FindByPage(offset, limit)
}
func (r *SSLCertRepo) GetDeployConfig(ctx context.Context, db *gorm.DB, id int64) (*model.SslDeployConfig, error) {
	q := query.Use(db).SslDeployConfig
	return q.WithContext(ctx).Where(q.ID.Eq(id)).First()
}
func (r *SSLCertRepo) CreateDeployConfig(ctx context.Context, db *gorm.DB, row *model.SslDeployConfig) error {
	return query.Use(db).SslDeployConfig.WithContext(ctx).Create(row)
}
func (r *SSLCertRepo) SaveDeployConfig(ctx context.Context, db *gorm.DB, row *model.SslDeployConfig) error {
	q := query.Use(db).SslDeployConfig
	return q.WithContext(ctx).Where(q.ID.Eq(row.ID)).Save(row)
}
func (r *SSLCertRepo) DeleteDeployConfig(ctx context.Context, db *gorm.DB, row *model.SslDeployConfig) error {
	_, err := query.Use(db).SslDeployConfig.WithContext(ctx).Delete(row)
	return err
}

func (r *SSLCertRepo) ListTasks(ctx context.Context, db *gorm.DB, cond *define.SSLTaskListCond) ([]*model.SslTask, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).SslTask
	do := q.WithContext(ctx)
	if cond.CertificateID > 0 {
		do = do.Where(q.CertificateID.Eq(cond.CertificateID))
	}
	if cond.TaskType != "" {
		do = do.Where(q.TaskType.Eq(cond.TaskType))
	}
	if cond.Status != "" {
		do = do.Where(q.Status.Eq(cond.Status))
	}
	if cond.TriggerType != "" {
		do = do.Where(q.TriggerType.Eq(cond.TriggerType))
	}
	return do.Order(q.ID.Desc()).FindByPage(offset, limit)
}
func (r *SSLCertRepo) GetTask(ctx context.Context, db *gorm.DB, id int64) (*model.SslTask, error) {
	q := query.Use(db).SslTask
	return q.WithContext(ctx).Where(q.ID.Eq(id)).First()
}
func (r *SSLCertRepo) CreateTask(ctx context.Context, db *gorm.DB, row *model.SslTask) error {
	return query.Use(db).SslTask.WithContext(ctx).Create(row)
}
func (r *SSLCertRepo) SaveTask(ctx context.Context, db *gorm.DB, row *model.SslTask) error {
	q := query.Use(db).SslTask
	return q.WithContext(ctx).Where(q.ID.Eq(row.ID)).Save(row)
}
