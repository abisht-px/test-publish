package migrations

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func UpdateImageEnvironments(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201910231045",
			Migrate: func(gdb *gorm.DB) (err error) {
				type Image struct {
					Environments string `gorm:"column:environments;type:text;not null;default:'unknown'"`
				}

				if err = gdb.Model(&Image{}).DropColumn("environment_id").Error; err != nil {
					return err
				}
				if err = gdb.Model(&Image{}).DropColumn("environment").Error; err != nil {
					return err
				}

				return gdb.AutoMigrate(&Image{}).Error
			},
			Rollback: func(gdb *gorm.DB) (err error) {
				type Image struct {
					EnvironmentID   uuid.UUID `gorm:"column:environment_id;type:uuid;not null;index"`
					EnvironmentName string    `gorm:"column:environment;type:text;not null"`
				}

				if err = gdb.Model(&Image{}).DropColumn("environments").Error; err != nil {
					return err
				}

				return gdb.AutoMigrate(&Image{}).Error
			},
		},
	})

	return m.Migrate()
}
