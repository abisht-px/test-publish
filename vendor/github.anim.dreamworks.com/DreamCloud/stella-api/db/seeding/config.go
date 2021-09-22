package seeding

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type Model struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;primary_key;not null"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamp;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamp;not null"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:timestamp;"`
}

type DatabaseType struct {
	Model
	Name                 string     `json:"name"`
	ShortName            string     `json:"short_name"`
	CanRegisterInVault   *bool      `json:"can_register_in_vault"`
	HasIncrementalBackup *bool      `json:"has_incremental_backup"`
	HasFullBackup        *bool      `json:"has_full_backup"`
	DeploymentCount      int        `json:"deployment_count" gorm:"-"` // Deployment Count is calculated by the API
	Versions             []Version  `json:"-" gorm:"-"`
	Templates            []Template `json:"-" gorm:"-"`

	DomainID uuid.UUID `json:"domain_id"`
}

type Version struct {
	Model
	Name             string    `json:"name"`
	DeploymentCount  int       `json:"deployment_count" gorm:"-"` // Deployment Count is cacluated by the API
	DatabaseTypeID   uuid.UUID `json:"database_type_id"`
	DatabaseTypeName string    `json:"database_type" gorm:"column:database_type"`
	Images           []Image   `json:"-" gorm:"-"`

	DomainID uuid.UUID `json:"domain_id"`
}

type Image struct {
	Model
	Registry         string    `json:"registry"`
	Namespace        string    `json:"namespace"`
	Name             string    `json:"name"`
	Tag              string    `json:"tag"`
	Build            string    `json:"build"`
	Environments     string    `json:"environments"`
	DeploymentCount  int       `json:"deployment_count" gorm:"-"`
	DatabaseTypeID   uuid.UUID `json:"database_type_id"`
	VersionID        uuid.UUID `json:"version_id"`
	DatabaseTypeName string    `json:"database_type" gorm:"column:database_type"`
	VersionName      string    `json:"version" gorm:"column:version"`

	DomainID uuid.UUID `json:"domain_id"`
}

type Template struct {
	Model
	SortOrder        uint                   `json:"sort_order"` // Allows you to sort template results (i.e. Small, Medium, Large)
	Name             string                 `json:"name"`
	DatabaseTypeName string                 `json:"database_type" gorm:"column:database_type"`
	DatabaseTypeID   uuid.UUID              `json:"database_type_id"`
	Resources        map[string]interface{} `json:"resources" gorm:"-"`
	Configurations   []Configuration        `json:"configurations" gorm:"-"`

	DomainID uuid.UUID `json:"domain_id"`

	ResourcesJSON        postgres.Jsonb `json:"-" gorm:"column:resources"`      // Used for DB
	ConfigurationsString string         `json:"-" gorm:"column:configurations"` // Used for DB
}

type Configuration struct {
	Name         string      `json:"name"`
	DefaultValue interface{} `json:"default_value"`
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
}

//================================================================================
// Hooks
//================================================================================

// BeforeCreate sets a new UUID on model creation.
// This hook is auto called by GORM on creation of the Model struct.
func (m *Model) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("ID", uuid.New())
}

func (template *Template) BeforeSave() (err error) {
	// Marhsal JSON
	if err = template.marshalJSON(); err != nil {
		return err
	}

	// Serialize Configurations
	if err = template.serializeConfigurations(); err != nil {
		return err
	}

	return nil
}

func (template *Template) marshalJSON() (err error) {
	// Resources JSON
	var resources []byte
	resources, err = json.Marshal(template.Resources)
	if err != nil {
		return err
	}
	template.ResourcesJSON = postgres.Jsonb{resources}

	return nil
}

func (template *Template) serializeConfigurations() (err error) {
	var configsArray []string

	for _, c := range template.Configurations {
		var config []byte
		config, err = json.Marshal(c)
		if err != nil {
			return err
		}
		configsArray = append(configsArray, string(config))
	}

	if len(configsArray) > 0 {
		template.ConfigurationsString = strings.Join(configsArray[:], ";;")
	}

	return nil
}
