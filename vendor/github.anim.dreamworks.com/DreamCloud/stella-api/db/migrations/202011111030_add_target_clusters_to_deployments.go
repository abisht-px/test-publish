package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddTargetClustersToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202011111030",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					TargetClusters string `gorm:"column:target_clusters;type:text"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("target_clusters").Error
			},
		},
	})

	return m.Migrate()
}
