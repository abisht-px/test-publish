package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddOriginToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202002101721",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					Origin string `gorm:"column:origin;type:text"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("origin").Error
			},
		},
	})

	return m.Migrate()
}
