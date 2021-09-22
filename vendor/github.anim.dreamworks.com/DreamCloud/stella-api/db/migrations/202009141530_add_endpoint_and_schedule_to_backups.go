package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AndEndpointAndScheduleToBackups(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202009141530",
			Migrate: func(gdb *gorm.DB) error {
				type Backup struct {
					Endpoint string `gorm:"column:endpoint;type:text"`
					Schedule string `gorm:"column:schedule;type:text"`
				}
				return gdb.AutoMigrate(&Backup{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Backup struct{}
				if err := gdb.Model(&Backup{}).DropColumn("endpoint").Error; err != nil {
					return err
				}
				return gdb.Model(&Backup{}).DropColumn("schedule").Error
			},
		},
	})

	return m.Migrate()
}
