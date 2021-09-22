package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddEndpointToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201910011533",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					Endpoint string `gorm:"column:endpoint;type:text"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("endpoint").Error
			},
		},
	})

	return m.Migrate()
}
