package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddStorageProviderToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201910221423",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					StorageProvider string `gorm:"column:storage_provider;type:text;not null;default:'unknown'"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("storage_provider").Error
			},
		},
	})

	return m.Migrate()
}
