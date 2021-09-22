package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddSortOrderToTemplates(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201910031626",
			Migrate: func(gdb *gorm.DB) error {
				type Template struct {
					SortOrder uint `gorm:"column:sort_order;type:smallint;not null;default:0"`
				}
				return gdb.AutoMigrate(&Template{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Template struct{}
				return gdb.Model(&Template{}).DropColumn("sort_order").Error
			},
		},
	})

	return m.Migrate()
}
