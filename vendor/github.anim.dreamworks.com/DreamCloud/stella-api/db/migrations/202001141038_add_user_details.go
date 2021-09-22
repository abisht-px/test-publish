package migrations

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddUserDetails(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202001141038",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					UserID    uuid.UUID `gorm:"column:user_id;type:uuid"`
					UserLogin string    `gorm:"column:user_login;type:text"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Deployment struct{}

				if err = gdb.Model(&Deployment{}).DropColumn("user_id").Error; err != nil {
					return err
				}

				return gdb.Model(&Deployment{}).DropColumn("user_login").Error
			},
		},
		{
			ID: "202001141039",
			Migrate: func(gdb *gorm.DB) error {
				type Snapshot struct {
					UserID    uuid.UUID `gorm:"column:user_id;type:uuid"`
					UserLogin string    `gorm:"column:user_login;type:text"`
				}
				return gdb.AutoMigrate(&Snapshot{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Snapshot struct{}

				if err = gdb.Model(&Snapshot{}).DropColumn("user_id").Error; err != nil {
					return err
				}

				return gdb.Model(&Snapshot{}).DropColumn("user_login").Error
			},
		},
	})

	return m.Migrate()
}
