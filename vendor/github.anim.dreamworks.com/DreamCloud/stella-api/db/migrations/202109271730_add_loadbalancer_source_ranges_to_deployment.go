package migrations

import (
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"gopkg.in/gormigrate.v1"
)

func AddLoadBalancerSourceRangesToDeployment(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202109271730",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					LoadBalancerSourceRanges postgres.Jsonb `gorm:"column:load_balancer_source_ranges;type:jsonb"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("load_balancer_source_ranges").Error
			},
		},
	})

	return m.Migrate()
}
