package migrations

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func UpdateEnvironmentClusters(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202005181500",
			Migrate: func(gdb *gorm.DB) error {
				type Environment struct {
					Clusters string `gorm:"column:clusters;type:text"`
				}

				if err := gdb.Model(&Environment{}).DropColumn("cluster_id").Error; err != nil {
					return err
				}
				if err := gdb.Model(&Environment{}).DropColumn("cluster_name").Error; err != nil {
					return err
				}
				if err := gdb.Model(&Environment{}).DropColumn("location").Error; err != nil {
					return err
				}

				return gdb.AutoMigrate(&Environment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Environment struct {
					Location    string    `gorm:"column:location;type:text;not null"`
					ClusterID   uuid.UUID `gorm:"column:cluster_id;type:uuid"`
					ClusterName string    `gorm:"column:cluster_name;type:text"`
				}

				if err := gdb.Model(&Environment{}).DropColumn("clusters").Error; err != nil {
					return err
				}

				return gdb.AutoMigrate(&Environment{}).Error
			},
		},
	})

	return m.Migrate()
}
