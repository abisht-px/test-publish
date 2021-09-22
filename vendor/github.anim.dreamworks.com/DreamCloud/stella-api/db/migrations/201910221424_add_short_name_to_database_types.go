package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddShortNameToDatabaseTypes(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201910221424",
			Migrate: func(gdb *gorm.DB) error {
				type DatabaseType struct {
					ShortName string `gorm:"column:short_name;type:text;not null;default:'unknown'"`
				}
				return gdb.AutoMigrate(&DatabaseType{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type DatabaseType struct{}
				return gdb.Model(&DatabaseType{}).DropColumn("short_name").Error
			},
		},
	})

	return m.Migrate()
}
