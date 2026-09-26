package biz

import (
	"gops/internal/pkg/utils"

	"gorm.io/gorm"
)

func queryPage[T any](query *gorm.DB, page, limit int) ([]T, int64, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset, limit := utils.PageBounds(page, limit)
	rows := make([]T, 0, limit)
	if err := query.Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func slicePage[T any](rows []T, page, limit int) ([]T, int64) {
	total := int64(len(rows))
	offset, limit := utils.PageBounds(page, limit)
	if offset >= len(rows) {
		return []T{}, total
	}
	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}
	return rows[offset:end], total
}
