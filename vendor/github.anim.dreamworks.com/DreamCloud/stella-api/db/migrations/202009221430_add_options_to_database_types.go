package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddOptionsToDatabaseTypes(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202009221430",
			Migrate: func(gdb *gorm.DB) error {
				type DatabaseType struct {
					CanRegisterInVault   bool `gorm:"column:can_register_in_vault;type:bool"`
					HasIncrementalBackup bool `gorm:"column:has_incremental_backup;type:bool"`
					HasFullBackup        bool `gorm:"column:has_full_backup;type:bool"`
				}
				return gdb.AutoMigrate(&DatabaseType{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type DatabaseType struct{}
				if err := gdb.Model(&DatabaseType{}).DropColumn("can_register_in_vault").Error; err != nil {
					return err
				}
				if err := gdb.Model(&DatabaseType{}).DropColumn("has_incremental_backup").Error; err != nil {
					return err
				}
				return gdb.Model(&DatabaseType{}).DropColumn("has_full_backup").Error
			},
		},
	})

	return m.Migrate()
}
