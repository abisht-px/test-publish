package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddServiceTypeParameterToDeployment(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202109171540",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					ServiceType string `gorm:"column:service_type;type:text"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("service_type").Error
			},
		},
	})

	return m.Migrate()
}
