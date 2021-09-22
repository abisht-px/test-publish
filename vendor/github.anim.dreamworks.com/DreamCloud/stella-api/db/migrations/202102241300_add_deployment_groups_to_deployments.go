package migrations

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddDeploymentGroupsToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202102241300",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					DeploymentGroupID   uuid.UUID `gorm:"column:deployment_group_id;type:uuid"`
					DeploymentGroupName string    `gorm:"column:deployment_group;type:text;default:'';not null"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Deployment struct{}
				if err := gdb.Model(&Deployment{}).DropColumn("deployment_group_id").Error; err != nil {
					return err
				}
				return gdb.Model(&Deployment{}).DropColumn("deployment_group").Error
			},
		},
	})

	return m.Migrate()
}
