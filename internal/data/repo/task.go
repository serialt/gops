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

type TaskRepo struct{}

func NewTaskRepo() *TaskRepo { return &TaskRepo{} }

func (r *TaskRepo) ListJobs(ctx context.Context, db *gorm.DB, cond *define.TaskJobListCond) ([]*model.TaskJob, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).TaskJob
	do := q.WithContext(ctx)
	if cond.Keyword != "" {
		do = do.Where(q.Name.Like("%" + cond.Keyword + "%"))
	}
	if cond.TaskType != "" {
		do = do.Where(q.TaskType.Eq(cond.TaskType))
	}
	if cond.Status != "" {
		do = do.Where(q.Status.Eq(cond.Status))
	}
	return do.Order(q.ID.Desc()).FindByPage(offset, limit)
}

func (r *TaskRepo) CreateJob(ctx context.Context, db *gorm.DB, row *model.TaskJob) error {
	return query.Use(db).TaskJob.WithContext(ctx).Create(row)
}

func (r *TaskRepo) UpdateJobResult(ctx context.Context, db *gorm.DB, id int64, status, result, errorMessage string) error {
	q := query.Use(db).TaskJob
	_, err := q.WithContext(ctx).Where(q.ID.Eq(id)).Updates(map[string]any{
		"status": status, "result": result, "error_message": errorMessage,
	})
	return err
}

func (r *TaskRepo) ListTemplates(ctx context.Context, db *gorm.DB, cond *define.TaskTemplateListCond) ([]*model.TaskTemplate, int64, error) {
	offset, limit := utils.PageBounds(cond.Page, cond.Limit)
	q := query.Use(db).TaskTemplate
	do := q.WithContext(ctx)
	if cond.Keyword != "" {
		value := "%" + cond.Keyword + "%"
		do = do.Where(field.Or(q.Name.Like(value), q.Code.Like(value)))
	}
	if cond.Category != "" {
		do = do.Where(q.Category.Eq(cond.Category))
	}
	if cond.Platform != "" {
		do = do.Where(q.Platform.Eq(cond.Platform))
	}
	if cond.Status != nil {
		do = do.Where(q.Status.Eq(*cond.Status))
	}
	return do.Order(q.SortOrder, q.ID.Desc()).FindByPage(offset, limit)
}

func (r *TaskRepo) GetTemplate(ctx context.Context, db *gorm.DB, id int64) (*model.TaskTemplate, error) {
	q := query.Use(db).TaskTemplate
	return q.WithContext(ctx).Where(q.ID.Eq(id)).First()
}

func (r *TaskRepo) CreateTemplate(ctx context.Context, db *gorm.DB, row *model.TaskTemplate) error {
	return query.Use(db).TaskTemplate.WithContext(ctx).Create(row)
}

func (r *TaskRepo) SaveTemplate(ctx context.Context, db *gorm.DB, row *model.TaskTemplate) error {
	q := query.Use(db).TaskTemplate
	return q.WithContext(ctx).Where(q.ID.Eq(row.ID)).Save(row)
}

func (r *TaskRepo) DeleteTemplate(ctx context.Context, db *gorm.DB, row *model.TaskTemplate) error {
	_, err := query.Use(db).TaskTemplate.WithContext(ctx).Delete(row)
	return err
}
