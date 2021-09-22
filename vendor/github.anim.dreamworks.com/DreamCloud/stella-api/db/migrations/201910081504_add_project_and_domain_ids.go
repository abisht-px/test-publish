package migrations

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

func AddProjectAndDomainIDs(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201910081504",
			Migrate: func(gdb *gorm.DB) error {
				type DatabaseType struct {
					DomainID uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}
				return gdb.AutoMigrate(&DatabaseType{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type DatabaseType struct{}
				return gdb.Model(&DatabaseType{}).DropColumn("domain_id").Error
			},
		},
		{
			ID: "201910081505",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					ProjectID uuid.UUID `gorm:"column:project_id;type:uuid;not null;index"`
					DomainID  uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Deployment struct{}
				if err := gdb.Model(&Deployment{}).DropColumn("project_id").Error; err != nil {
					return err
				}
				return gdb.Model(&Deployment{}).DropColumn("domain_id").Error
			},
		},
		{
			ID: "201910081506",
			Migrate: func(gdb *gorm.DB) error {
				type Environment struct {
					DomainID uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}
				return gdb.AutoMigrate(&Environment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Environment struct{}
				return gdb.Model(&Environment{}).DropColumn("domain_id").Error
			},
		},
		{
			ID: "201910081507",
			Migrate: func(gdb *gorm.DB) error {
				type Image struct {
					DomainID uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}
				return gdb.AutoMigrate(&Image{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Image struct{}
				return gdb.Model(&Image{}).DropColumn("domain_id").Error
			},
		},
		{
			ID: "201910081508",
			Migrate: func(gdb *gorm.DB) error {
				type Snapshot struct {
					ProjectID uuid.UUID `gorm:"column:project_id;type:uuid;not null;index"`
					DomainID  uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}
				return gdb.AutoMigrate(&Snapshot{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Snapshot struct{}
				if err := gdb.Model(&Snapshot{}).DropColumn("project_id").Error; err != nil {
					return err
				}
				return gdb.Model(&Snapshot{}).DropColumn("domain_id").Error
			},
		},
		{
			ID: "201910081509",
			Migrate: func(gdb *gorm.DB) error {
				type Task struct {
					ProjectID uuid.UUID `gorm:"column:project_id;type:uuid;not null;index"`
					DomainID  uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}
				return gdb.AutoMigrate(&Task{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Task struct{}
				if err := gdb.Model(&Task{}).DropColumn("project_id").Error; err != nil {
					return err
				}
				return gdb.Model(&Task{}).DropColumn("domain_id").Error
			},
		},
		{
			ID: "201910081510",
			Migrate: func(gdb *gorm.DB) error {
				type Version struct {
					DomainID uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}
				return gdb.AutoMigrate(&Version{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Version struct{}
				return gdb.Model(&Version{}).DropColumn("domain_id").Error
			},
		},
		{
			ID: "201910081511",
			Migrate: func(gdb *gorm.DB) error {
				type Template struct {
					DomainID uuid.UUID `gorm:"column:domain_id;type:uuid;not null;index"`
				}
				return gdb.AutoMigrate(&Template{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				type Template struct{}
				return gdb.Model(&Template{}).DropColumn("domain_id").Error
			},
		},
	})

	return m.Migrate()
}
