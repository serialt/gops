package data

import (
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(
	NewDB,
	NewData,
)

type Data struct {
	db *gorm.DB
}

func NewData(db *gorm.DB) *Data {
	return &Data{
		db: db,
	}
}

func (d *Data) GetDB() (db *gorm.DB) {
	return d.db
}
