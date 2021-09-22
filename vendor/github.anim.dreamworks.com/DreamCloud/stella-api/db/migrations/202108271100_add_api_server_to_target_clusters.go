package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddAPIServerToTargetClusters(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202108271100",
			Migrate: func(gdb *gorm.DB) error {
				type TargetCluster struct {
					APIServer string `gorm:"column:api_server;type:text"`
				}
				return gdb.AutoMigrate(&TargetCluster{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type TargetCluster struct{}
				return gdb.Model(&TargetCluster{}).DropColumn("api_server").Error
			},
		},
	})

	return m.Migrate()
}
