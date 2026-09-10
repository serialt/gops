package biz

import (
	"database/sql"
	"errors"

	"gops/internal/data/api"
	"gops/internal/data/model"

	"gorm.io/gorm"
)

type OrganizationBiz struct {
	db *gorm.DB
}

func NewOrganizationBiz(db *gorm.DB) *OrganizationBiz {
	return &OrganizationBiz{db: db}
}

func (b *OrganizationBiz) List(req *api.OrganizationQueryRequest) ([]*api.OrganizationResponse, error) {
	var rows []model.Organization
	query := b.db.Order("parent_id NULLS FIRST, sort_order ASC, id ASC")
	if req.Keyword != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]*api.OrganizationResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, toOrganizationResponse(row))
	}
	return result, nil
}

func (b *OrganizationBiz) Create(req *api.OrganizationRequest) (*api.OrganizationResponse, error) {
	if err := b.validateParent(req.ParentID, 0); err != nil {
		return nil, err
	}
	row := model.Organization{ParentID: parentIDValue(req.ParentID), Name: req.Name, Code: req.Code, Manager: req.Manager, Status: *req.Status}
	query := b.db
	if req.ParentID == nil || *req.ParentID == 0 {
		query = query.Omit("ParentID")
	}
	if err := query.Create(&row).Error; err != nil {
		return nil, err
	}
	return toOrganizationResponse(row), nil
}

func (b *OrganizationBiz) Update(req *api.OrganizationRequest) (*api.OrganizationResponse, error) {
	if req.ID == 0 {
		return nil, ErrInvalid
	}
	if err := b.validateParent(req.ParentID, req.ID); err != nil {
		return nil, err
	}
	var row model.Organization
	if err := b.db.First(&row, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if row.Code == "root" && req.Code != "root" {
		return nil, errors.New("root organization code cannot be changed")
	}
	var updateErr error
	if req.ParentID == nil || *req.ParentID == 0 {
		updateErr = b.db.Model(&row).Updates(map[string]any{
			"parent_id": nil, "name": req.Name, "code": req.Code, "manager": req.Manager, "status": *req.Status,
		}).Error
	} else {
		row.ParentID = parentIDValue(req.ParentID)
		row.Name = req.Name
		row.Code = req.Code
		row.Manager = req.Manager
		row.Status = *req.Status
		updateErr = b.db.Save(&row).Error
	}
	if updateErr != nil {
		return nil, updateErr
	}
	return toOrganizationResponse(row), nil
}

func (b *OrganizationBiz) Delete(req *api.OrganizationIDRequest) error {
	var organization model.Organization
	if err := b.db.First(&organization, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if organization.Code == "root" {
		return errors.New("root organization cannot be deleted")
	}
	var count int64
	if err := b.db.Model(&model.Organization{}).Where("parent_id = ?", req.ID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("remove child organizations first")
	}
	result := b.db.Delete(&model.Organization{}, req.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (b *OrganizationBiz) Move(req *api.OrganizationMoveRequest) (*api.OrganizationResponse, error) {
	if err := b.validateParent(req.ParentID, req.ID); err != nil {
		return nil, err
	}
	if b.wouldCreateCycle(req.ParentID, req.ID) {
		return nil, errors.New("organization cannot be moved under its descendant")
	}
	var row model.Organization
	if err := b.db.First(&row, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if row.Code == "root" {
		return nil, errors.New("root organization cannot be moved")
	}
	if err := b.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{"sort_order": req.SortOrder}
		if req.ParentID == nil || *req.ParentID == 0 {
			updates["parent_id"] = nil
		} else {
			updates["parent_id"] = *req.ParentID
		}
		if err := tx.Model(&row).Updates(updates).Error; err != nil {
			return err
		}

		siblings := tx.Where("id <> ?", row.ID).Order("sort_order ASC, id ASC")
		if req.ParentID == nil || *req.ParentID == 0 {
			siblings = siblings.Where("parent_id IS NULL")
		} else {
			siblings = siblings.Where("parent_id = ?", *req.ParentID)
		}
		var rows []model.Organization
		if err := siblings.Find(&rows).Error; err != nil {
			return err
		}
		position := int(req.SortOrder)
		if position < 0 {
			position = 0
		}
		if position > len(rows) {
			position = len(rows)
		}
		rows = append(rows, model.Organization{})
		copy(rows[position+1:], rows[position:])
		rows[position] = row
		for index, sibling := range rows {
			if err := tx.Model(&model.Organization{}).Where("id = ?", sibling.ID).Update("sort_order", int64(index)*10).Error; err != nil {
				return err
			}
		}
		row = rows[position]
		row.SortOrder = int64(position) * 10
		return nil
	}); err != nil {
		return nil, err
	}
	return toOrganizationResponse(row), nil
}

func (b *OrganizationBiz) wouldCreateCycle(parentID *int64, self int64) bool {
	if parentID == nil || *parentID == 0 {
		return false
	}
	current := *parentID
	seen := map[int64]struct{}{}
	for current != 0 {
		if current == self {
			return true
		}
		if _, exists := seen[current]; exists {
			return true
		}
		seen[current] = struct{}{}
		var parent sql.NullInt64
		if err := b.db.Model(&model.Organization{}).Select("parent_id").Where("id = ?", current).Scan(&parent).Error; err != nil {
			return true
		}
		if !parent.Valid {
			return false
		}
		current = parent.Int64
	}
	return false
}

func (b *OrganizationBiz) validateParent(parentID *int64, self int64) error {
	if parentID == nil || *parentID == 0 {
		return nil
	}
	if *parentID == self {
		return errors.New("organization cannot be its own parent")
	}
	var count int64
	if err := b.db.Model(&model.Organization{}).Where("id = ?", *parentID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("parent organization not found")
	}
	return nil
}

func parentIDValue(parentID *int64) int64 {
	if parentID == nil {
		return 0
	}
	return *parentID
}

func toOrganizationResponse(row model.Organization) *api.OrganizationResponse {
	var parentID *int64
	if row.ParentID != 0 {
		parentID = &row.ParentID
	}
	return &api.OrganizationResponse{ID: row.ID, ParentID: parentID, SortOrder: row.SortOrder, Name: row.Name, Code: row.Code, Manager: row.Manager, Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}
