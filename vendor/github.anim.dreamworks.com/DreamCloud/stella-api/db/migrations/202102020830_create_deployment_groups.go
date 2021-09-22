package migrations

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func CreateDeploymentGroups(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202102020830",
			Migrate: func(gdb *gorm.DB) error {
				type DeploymentGroup struct {
					Model
					Name     string    `gorm:"column:name;type:text"`
					DomainID uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}
				return gdb.AutoMigrate(&DeploymentGroup{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				return gdb.DropTable("deployment_groups").Error
			},
		},
	})

	return m.Migrate()
}
