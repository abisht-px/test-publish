package migrations

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddClusterIDToEnvironments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201910011605",
			Migrate: func(gdb *gorm.DB) error {
				type Environment struct {
					ClusterID   uuid.UUID `gorm:"column:cluster_id;type:uuid"`
					ClusterName string    `gorm:"column:cluster_name;type:text"`
				}
				return gdb.AutoMigrate(&Environment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Environment struct{}
				if err := gdb.Model(&Environment{}).DropColumn("cluster_id").Error; err != nil {
					return err
				}
				return gdb.Model(&Environment{}).DropColumn("cluster_name").Error
			},
		},
	})

	return m.Migrate()
}
