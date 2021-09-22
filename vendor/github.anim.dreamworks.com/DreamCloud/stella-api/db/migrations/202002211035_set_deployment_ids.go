package migrations

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.anim.dreamworks.com/DreamCloud/stella-api/utils"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"gopkg.in/gormigrate.v1"
)

func SetDeploymentIDs(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202002211035",
			Migrate: func(gdb *gorm.DB) error {
				type Deployment struct {
					Model
					NodeCount        uint       `json:"node_count" mapstructure:"node_count"`
					DeploymentID     string     `json:"deployment_id" mapstructure:"deployment_id"`
					ClusterName      string     `json:"cluster_name" mapstructure:"cluster_name"`
					Build            string     `json:"build" mapstructure:"build"`
					Origin           string     `json:"origin" mapstructure:"-"`
					State            string     `json:"state" mapstructure:"-"`
					Schema           string     `json:"schema" mapstructure:"schema"`
					Endpoint         string     `json:"endpoint" mapstructure:"-"`
					Service          string     `json:"service" mapstructure:"service"`
					StorageProvider  string     `json:"storage_provider" mapstructure:"storage_provider"`
					LastBackup       *time.Time `json:"last_backup" mapstructure:"-" binding:"-"`
					DatabaseTypeID   uuid.UUID  `json:"database_type_id" mapstructure:"-"`
					VersionID        uuid.UUID  `json:"version_id" mapstructure:"-"`
					ImageID          *uuid.UUID `json:"image_id" mapstructure:"-"`
					EnvironmentID    uuid.UUID  `json:"environment_id" mapstructure:"-"`
					DatabaseTypeName string     `json:"database_type" mapstructure:"database_type" gorm:"column:database_type"`
					VersionName      string     `json:"version" mapstructure:"version" gorm:"column:version"`
					ImageName        string     `json:"image" mapstructure:"image" gorm:"column:image"`
					EnvironmentName  string     `json:"environment" mapstructure:"environment" gorm:"column:environment"`

					DomainID  uuid.UUID `json:"domain_id" mapstructure:"-"`
					ProjectID uuid.UUID `json:"project_id" mapstructure:"-"`
					UserID    uuid.UUID `json:"user_id" mapstructure:"-"`
					UserLogin string    `json:"user_login" mapstructure:"-"`

					ConfigurationJSON     postgres.Jsonb `json:"-" mapstructure:"-" gorm:"column:configuration"`      // Used for DB
					ResourcesJSON         postgres.Jsonb `json:"-" mapstructure:"-" gorm:"column:resources"`          // Used for DB
					ConnectionDetailsJSON postgres.Jsonb `json:"-" mapstructure:"-" gorm:"column:connection_details"` // Used for DB
				}

				var deployments []Deployment
				gdb.Unscoped().Find(&deployments)

				for _, deployment := range deployments {
					if deployment.Build == "" {
						// Generate build (loop to ensure build is unique)
						build := ""
						for {
							seed := fmt.Sprintf("%v-%v-%v", deployment.ClusterName, deployment.EnvironmentName, time.Now())
							sha := fmt.Sprintf("%x", sha256.Sum256([]byte(seed)))
							build = utils.Substring(sha, 0, 8)

							var count int
							db.Unscoped().Model(&Deployment{}).Where("build = ?", build).Count(&count)
							if count == 0 {
								break // build is unique
							}
						}

						// Update Build and DeploymentID
						resp := gdb.Model(&deployment).Update(map[string]interface{}{
							"build":         build,
							"deployment_id": fmt.Sprintf("%v-%v", deployment.ClusterName, build),
						})

						if resp.Error != nil {
							return resp.Error
						}
					}
				}

				return nil
			},
		},
	})

	return m.Migrate()
}
