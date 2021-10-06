package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddComingSoonToDatabaseTypes(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202109200930",
			Migrate: func(gdb *gorm.DB) error {
				type DatabaseType struct {
					ComingSoon bool `gorm:"column:coming_soon;type:bool;not null;default:false"`
				}
				return gdb.AutoMigrate(&DatabaseType{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type DatabaseType struct{}
				return gdb.Model(&DatabaseType{}).DropColumn("coming_soon").Error
			},
		},
	})

	return m.Migrate()
}
