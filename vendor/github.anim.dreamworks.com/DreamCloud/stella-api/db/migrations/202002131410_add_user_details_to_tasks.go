package migrations

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddUserDetailsToTasks(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202002131410",
			Migrate: func(gdb *gorm.DB) error {
				type Task struct {
					UserID    uuid.UUID `gorm:"column:user_id;type:uuid"`
					UserLogin string    `gorm:"column:user_login;type:text"`
				}
				return gdb.AutoMigrate(&Task{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Task struct{}

				if err = gdb.Model(&Task{}).DropColumn("user_id").Error; err != nil {
					return err
				}

				return gdb.Model(&Task{}).DropColumn("user_login").Error
			},
		},
	})

	return m.Migrate()
}
