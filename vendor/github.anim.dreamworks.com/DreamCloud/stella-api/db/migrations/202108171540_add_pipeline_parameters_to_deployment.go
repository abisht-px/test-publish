package migrations

import (
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"gopkg.in/gormigrate.v1"
)

func AddPipelineParametersToDeployment(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202108171540",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					Initialize              string         `gorm:"column:initialize;type:text"`
					ImagePullSecret         string         `gorm:"column:image_pull_secret;type:text"`
					DNSZone                 string         `gorm:"column:dns_zone;type:text"`
					StorageClassProvisioner string         `gorm:"column:storage_class_provisioner;type:text"`
					StorageOptions          postgres.Jsonb `gorm:"column:storage_options;type:jsonb"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Deployment struct{}
				if err := gdb.Model(&Deployment{}).DropColumn("initialize").Error; err != nil {
					return err
				}
				if err := gdb.Model(&Deployment{}).DropColumn("image_pull_secret").Error; err != nil {
					return err
				}
				if err := gdb.Model(&Deployment{}).DropColumn("dns_zone").Error; err != nil {
					return err
				}
				if err := gdb.Model(&Deployment{}).DropColumn("storage_class_provisioner").Error; err != nil {
					return err
				}
				return gdb.Model(&Deployment{}).DropColumn("storage_options").Error
			},
		},
	})

	return m.Migrate()
}
