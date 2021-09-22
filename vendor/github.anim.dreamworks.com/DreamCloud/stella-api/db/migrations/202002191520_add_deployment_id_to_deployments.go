package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddDeploymentIDToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202002191520",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					DeploymentID string `gorm:"column:deployment_id;type:text"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("deployment_id").Error
			},
		},
	})

	return m.Migrate()
}
