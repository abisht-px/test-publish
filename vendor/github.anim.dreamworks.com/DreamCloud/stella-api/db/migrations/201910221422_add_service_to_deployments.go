package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddServiceToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201910221422",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					Service string `gorm:"column:service;type:text;not null;default:'unknown'"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("service").Error
			},
		},
	})

	return m.Migrate()
}
