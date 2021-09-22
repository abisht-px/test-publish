package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddRegisteredInVaultToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202004011615",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					RegisteredInVault bool `gorm:"column:registered_in_vault;type:bool"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("registered_in_vault").Error
			},
		},
	})

	return m.Migrate()
}
