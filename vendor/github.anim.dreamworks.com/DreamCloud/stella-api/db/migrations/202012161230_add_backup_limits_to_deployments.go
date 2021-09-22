package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddBackupLimitsToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202012161230",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					FullBackupLimit        uint `gorm:"column:full_backup_limit;type:smallint"`
					IncrementalBackupLimit uint `gorm:"column:incremental_backup_limit;type:smallint"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Deployment struct{}
				if err := gdb.Model(&Deployment{}).DropColumn("full_backup_limit").Error; err != nil {
					return err
				}
				return gdb.Model(&Deployment{}).DropColumn("incremental_backup_limit").Error
			},
		},
	})

	return m.Migrate()
}
