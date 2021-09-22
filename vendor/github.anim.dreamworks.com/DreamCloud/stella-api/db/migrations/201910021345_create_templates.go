package migrations

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"gopkg.in/gormigrate.v1"
)

func CreateTemplates(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201910021345",
			Migrate: func(gdb *gorm.DB) error {
				type Template struct {
					Model
					Name             string         `gorm:"column:name;type:text;not null"`
					DatabaseTypeName string         `gorm:"column:database_type;type:text;not null"`
					DatabaseTypeID   uuid.UUID      `gorm:"column:database_type_id;type:uuid;not null"`
					Resources        postgres.Jsonb `gorm:"column:resources;type:jsonb;not null"`
				}
				return gdb.AutoMigrate(&Template{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				return gdb.DropTable("templates").Error
			},
		},
	})

	return m.Migrate()
}
