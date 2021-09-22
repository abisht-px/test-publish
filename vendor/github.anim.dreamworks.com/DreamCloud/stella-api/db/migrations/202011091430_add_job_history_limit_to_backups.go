package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddJobHistoryLimitToBackups(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202011091430",
			Migrate: func(gdb *gorm.DB) error {
				type Backup struct {
					JobHistoryLimit uint `gorm:"column:job_history_limit;type:smallint"`
				}
				return gdb.AutoMigrate(&Backup{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Backup struct{}
				return gdb.Model(&Backup{}).DropColumn("job_history_limit").Error
			},
		},
	})

	return m.Migrate()
}
