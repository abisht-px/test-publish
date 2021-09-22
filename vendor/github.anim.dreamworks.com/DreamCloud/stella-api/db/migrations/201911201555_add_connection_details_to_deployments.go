package migrations

import (
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"gopkg.in/gormigrate.v1"
)

func AddConnectionDetailsToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201911201555",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					ConnectionDetails postgres.Jsonb `gorm:"column:connection_details;type:jsonb"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("connection_details").Error
			},
		},
	})

	return m.Migrate()
}
