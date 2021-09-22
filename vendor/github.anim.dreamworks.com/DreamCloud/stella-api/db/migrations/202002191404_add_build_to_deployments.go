package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddBuildToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202002191404",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					Build string `gorm:"column:build;type:text"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("build").Error
			},
		},
	})

	return m.Migrate()
}
