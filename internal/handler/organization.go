package handler

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"gops/internal/data/model"
	"gops/internal/pkg/response"
	"gorm.io/gorm"
)

type OrganizationHandler struct{ db *gorm.DB }

func NewOrganizationHandler(db *gorm.DB) *OrganizationHandler { return &OrganizationHandler{db: db} }

type organizationRequest struct {
	ID       int64  `json:"id"`
	ParentID *int64 `json:"parent_id"`
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Manager  string `json:"manager"`
	Status   *bool  `json:"status" binding:"required"`
}
type organizationMoveRequest struct {
	ID        int64  `json:"id" binding:"required"`
	ParentID  *int64 `json:"parent_id"`
	SortOrder int64  `json:"sort_order"`
}

func (h *OrganizationHandler) List(c *gin.Context) {
	var req struct {
		Keyword string `json:"keyword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid organization query")
		return
	}
	var rows []model.Organization
	q := h.db.Order("parent_id NULLS FIRST, sort_order ASC, id ASC")
	if req.Keyword != "" {
		q = q.Where("name ILIKE ? OR code ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if err := q.Find(&rows).Error; err != nil {
		response.Error(c, 500, "query organizations failed")
		return
	}
	response.OK(c, rows)
}
func (h *OrganizationHandler) Create(c *gin.Context) {
	var req organizationRequest
	if c.ShouldBindJSON(&req) != nil {
		response.Error(c, 400, "invalid organization data")
		return
	}
	if err := h.validateParent(req.ParentID, 0); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	row := model.Organization{ParentID: parentIDValue(req.ParentID), Name: req.Name, Code: req.Code, Manager: req.Manager, Status: *req.Status}
	query := h.db
	if req.ParentID == nil || *req.ParentID == 0 {
		query = query.Omit("ParentID")
	}
	if err := query.Create(&row).Error; err != nil {
		response.Error(c, 400, "create organization failed: "+err.Error())
		return
	}
	response.Created(c, row)
}
func (h *OrganizationHandler) Update(c *gin.Context) {
	var req organizationRequest
	if c.ShouldBindJSON(&req) != nil {
		response.Error(c, 400, "invalid organization data")
		return
	}
	if req.ID == 0 {
		response.Error(c, 400, "invalid organization id")
		return
	}
	if err := h.validateParent(req.ParentID, req.ID); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	var row model.Organization
	if err := h.db.First(&row, req.ID).Error; err != nil {
		response.Error(c, 404, "organization not found")
		return
	}
	if row.Code == "root" && req.Code != "root" {
		response.Error(c, 400, "root organization code cannot be changed")
		return
	}
	var updateErr error
	if req.ParentID == nil || *req.ParentID == 0 {
		updateErr = h.db.Model(&row).Updates(map[string]any{
			"parent_id": nil, "name": req.Name, "code": req.Code, "manager": req.Manager, "status": *req.Status,
		}).Error
	} else {
		row.ParentID = parentIDValue(req.ParentID)
		row.Name = req.Name
		row.Code = req.Code
		row.Manager = req.Manager
		row.Status = *req.Status
		updateErr = h.db.Save(&row).Error
	}
	if updateErr != nil {
		response.Error(c, 400, "update organization failed: "+updateErr.Error())
		return
	}
	response.OK(c, row)
}
func (h *OrganizationHandler) Delete(c *gin.Context) {
	var req struct {
		ID int64 `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid organization id")
		return
	}
	id := req.ID
	var count int64
	var organization model.Organization
	if err := h.db.First(&organization, id).Error; err != nil {
		response.Error(c, 404, "organization not found")
		return
	}
	if organization.Code == "root" {
		response.Error(c, 400, "root organization cannot be deleted")
		return
	}
	h.db.Model(&model.Organization{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		response.Error(c, 400, "remove child organizations first")
		return
	}
	r := h.db.Delete(&model.Organization{}, id)
	if r.Error != nil {
		response.Error(c, 500, "delete organization failed")
		return
	}
	if r.RowsAffected == 0 {
		response.Error(c, 404, "organization not found")
		return
	}
	response.OK(c, nil)
}

func (h *OrganizationHandler) Move(c *gin.Context) {
	var req organizationMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid move data")
		return
	}
	id := req.ID
	if err := h.validateParent(req.ParentID, id); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	if h.wouldCreateCycle(req.ParentID, id) {
		response.Error(c, 400, "organization cannot be moved under its descendant")
		return
	}
	var row model.Organization
	if err := h.db.First(&row, id).Error; err != nil {
		response.Error(c, 404, "organization not found")
		return
	}
	if row.Code == "root" {
		response.Error(c, 400, "root organization cannot be moved")
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
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
		response.Error(c, 500, "move organization failed")
		return
	}
	response.OK(c, row)
}

func (h *OrganizationHandler) wouldCreateCycle(parentID *int64, self int64) bool {
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
		if err := h.db.Model(&model.Organization{}).Select("parent_id").Where("id = ?", current).Scan(&parent).Error; err != nil {
			return true
		}
		if !parent.Valid {
			return false
		}
		current = parent.Int64
	}
	return false
}
func (h *OrganizationHandler) validateParent(parentID *int64, self int64) error {
	if parentID == nil || *parentID == 0 {
		return nil
	}
	if *parentID == self {
		return errors.New("organization cannot be its own parent")
	}
	var count int64
	h.db.Model(&model.Organization{}).Where("id = ?", *parentID).Count(&count)
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
