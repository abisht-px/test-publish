package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AndTypeAndLevelToBackups(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202008261630",
			Migrate: func(gdb *gorm.DB) error {
				type Backup struct {
					BackupType  string `gorm:"column:backup_type;type:text"`
					BackupLevel string `gorm:"column:backup_level;type:text"`
				}
				return gdb.AutoMigrate(&Backup{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Backup struct{}
				if err := gdb.Model(&Backup{}).DropColumn("backup_type").Error; err != nil {
					return err
				}
				return gdb.Model(&Backup{}).DropColumn("backup_level").Error
			},
		},
	})

	return m.Migrate()
}
