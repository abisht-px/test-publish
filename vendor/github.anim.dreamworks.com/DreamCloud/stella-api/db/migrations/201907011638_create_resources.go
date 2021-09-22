package migrations

import (
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"gopkg.in/gormigrate.v1"
)

func CreateResources(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "201907011638",
			Migrate: func(gdb *gorm.DB) error {
				type DatabaseType struct {
					Model
					Name string `gorm:"column:name;type:text;index;not null"`
				}
				return gdb.AutoMigrate(&DatabaseType{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				return gdb.DropTable("database_types").Error
			},
		},
		{
			ID: "201907011639",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					Model
					NodeCount        uint           `gorm:"column:node_count;type:smallint;not null"`
					ClusterName      string         `gorm:"column:cluster_name;type:text;index;not null"`
					State            string         `gorm:"column:state;type:text;not null"`
					Schema           string         `gorm:"column:schema;type:text;not null"`
					Configuration    postgres.Jsonb `gorm:"column:configuration;type:jsonb;not null"`
					Resources        postgres.Jsonb `gorm:"column:resources;type:jsonb;not null"`
					LastBackup       *time.Time     `gorm:"column:last_backup;type:timestamp"`
					DatabaseTypeID   uuid.UUID      `gorm:"column:database_type_id;type:uuid;not null;index"`
					VersionID        uuid.UUID      `gorm:"column:version_id;type:uuid;not null;index"`
					ImageID          *uuid.UUID     `gorm:"column:image_id;type:uuid"`
					EnvironmentID    uuid.UUID      `gorm:"column:environment_id;type:uuid;not null;index"`
					DatabaseTypeName string         `gorm:"column:database_type;type:text;not null"`
					VersionName      string         `gorm:"column:version;type:text;not null"`
					ImageName        string         `gorm:"column:image;type:text;not null"`
					EnvironmentName  string         `gorm:"column:environment;type:text;not null"`
				}
				return gdb.AutoMigrate(&Deployment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				return gdb.DropTable("deployments").Error
			},
		},
		{
			ID: "201907011640",
			Migrate: func(gdb *gorm.DB) error {
				type Environment struct {
					Model
					Name     string `gorm:"column:name;type:text;not null"`
					Location string `gorm:"column:location;type:text;not null"`
				}
				return gdb.AutoMigrate(&Environment{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				return gdb.DropTable("environments").Error
			},
		},
		{
			ID: "201907011641",
			Migrate: func(gdb *gorm.DB) error {
				type Image struct {
					Model
					Registry         string    `gorm:"column:registry;type:text;not null"`
					Namespace        string    `gorm:"column:namespace;type:text;not null"`
					Name             string    `gorm:"column:name;type:text;not null"`
					Tag              string    `gorm:"column:tag;type:text;not null"`
					Build            string    `gorm:"column:build;type:text;not null"`
					DatabaseTypeID   uuid.UUID `gorm:"column:database_type_id;type:uuid;not null;index"`
					VersionID        uuid.UUID `gorm:"column:version_id;type:uuid;not null;index"`
					EnvironmentID    uuid.UUID `gorm:"column:environment_id;type:uuid;not null;index"`
					DatabaseTypeName string    `gorm:"column:database_type;type:text;not null"`
					VersionName      string    `gorm:"column:version;type:text;not null"`
					EnvironmentName  string    `gorm:"column:environment;type:text;not null"`
				}
				return gdb.AutoMigrate(&Image{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				return gdb.DropTable("images").Error
			},
		},
		{
			ID: "201907011642",
			Migrate: func(gdb *gorm.DB) error {
				type Snapshot struct {
					Model
					FileSize       uint      `gorm:"column:file_size;type:integer;not null"`
					SnapshotTime   time.Time `gorm:"column:snapshot_time;type:timestamp;not null"`
					DeploymentID   uuid.UUID `gorm:"column:deployment_id;type:uuid;not null;index"`
					DeploymentName string    `gorm:"column:deployment;type:text;not null"`
				}
				return gdb.AutoMigrate(&Snapshot{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				return gdb.DropTable("snapshots").Error
			},
		},
		{
			ID: "201907011643",
			Migrate: func(gdb *gorm.DB) error {
				type Task struct {
					TaskModel
					Description        string `gorm:"column:description;type:text;not null"`
					TotalSteps         uint   `gorm:"column:total_steps;type:integer;not null"`
					CurrentStep        uint   `gorm:"column:current_step;type:integer;not null"`
					Status             string `gorm:"column:status;type:text;not null"`
					Log                string `gorm:"column:log;type:text;not null"`
					AssociatedResource string `gorm:"column:associated_resource;type:text"`
				}
				return gdb.AutoMigrate(&Task{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				return gdb.DropTable("tasks").Error
			},
		},
		{
			ID: "201907011644",
			Migrate: func(gdb *gorm.DB) error {
				type Version struct {
					Model
					Name             string    `gorm:"column:name;type:text;index;not null"`
					DatabaseTypeID   uuid.UUID `gorm:"column:database_type_id;type:uuid;not null;index"`
					DatabaseTypeName string    `gorm:"column:database_type;type:text;not null"`
				}
				return gdb.AutoMigrate(&Version{}).Error
			},
			Rollback: func(gdb *gorm.DB) error {
				return gdb.DropTable("versions").Error
			},
		},
	})

	return m.Migrate()
}
