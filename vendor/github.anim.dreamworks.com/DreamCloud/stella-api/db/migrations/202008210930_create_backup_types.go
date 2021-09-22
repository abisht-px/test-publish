package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func CreateBackupTypes(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202008210930",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					FullBackupSchedule        string `gorm:"column:full_backup_schedule;type:text"`
					IncrementalBackupSchedule string `gorm:"column:incremental_backup_schedule;type:text"`
				}
				if err := gdb.Model(&Deployment{}).DropColumn("backup_schedule").Error; err != nil {
					return err
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Deployment struct {
					BackupSchedule string `gorm:"column:backup_schedule;type:text"`
				}
				if err := gdb.Model(&Deployment{}).DropColumn("full_backup_schedule").Error; err != nil {
					return err
				}
				if err := gdb.Model(&Deployment{}).DropColumn("incremental_backup_schedule").Error; err != nil {
					return err
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
		},
	})

	return m.Migrate()
}
