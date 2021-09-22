package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddBackupIDToBackups(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202009211330",
			Migrate: func(gdb *gorm.DB) error {
				type Backup struct {
					Build    string `gorm:"column:build;type:text"`
					BackupID string `gorm:"column:backup_id;type:text"`
				}
				return gdb.AutoMigrate(&Backup{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Backup struct{}
				if err := gdb.Model(&Backup{}).DropColumn("build").Error; err != nil {
					return err
				}
				return gdb.Model(&Backup{}).DropColumn("backup_id").Error
			},
		},
	})

	return m.Migrate()
}
