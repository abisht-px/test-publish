package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddConfigurationsToTemplates(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201910231545",
			Migrate: func(gdb *gorm.DB) error {
				type Template struct {
					Configurations string `gorm:"column:configurations;type:text;not null;default:''"`
				}
				return gdb.AutoMigrate(&Template{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Template struct{}
				return gdb.Model(&Template{}).DropColumn("configurations").Error
			},
		},
	})

	return m.Migrate()
}
