package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddBackupScheduleToDeployments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202008130930",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					BackupSchedule string `gorm:"column:backup_schedule;type:text"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Deployment struct{}
				return gdb.Model(&Deployment{}).DropColumn("backup_schedule").Error
			},
		},
	})

	return m.Migrate()
}
