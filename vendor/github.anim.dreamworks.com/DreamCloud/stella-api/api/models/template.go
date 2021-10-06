package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type Configuration struct {
	Name         string      `json:"name"`
	DefaultValue interface{} `json:"default_value"`
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
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

func NewTemplate() Template {
	return Template{}
}

func AllTemplates() []Template {
	var templates []Template
	db.Unscoped().Find(&templates)
	return templates
}

func TemplatesCount(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
) int {
	var count int
	var template []Template

	sDB := dbWithScope(archived, domainIDs, nil)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("database_type LIKE ?", searchParam)
	}

	sDB.Find(&template).Count(&count)
	return count
}

func PaginatedTemplates(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
	page int,
	perPage int,
	orderBy string,
) []Template {
	var templates []Template
	offset := (page - 1) * perPage

	if orderBy == "" {
		orderBy = "created_at desc"
	}

	sDB := dbWithScope(archived, domainIDs, nil).Order(orderBy).Offset(offset).Limit(perPage).Where(queryParams)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("database_type LIKE ?", searchParam)
	}

	sDB.Find(&templates)
	return templates
}

func FindTemplate(id interface{}) (*Template, error) {
	template := NewTemplate()
	resp := Find(&template, id)

	if resp.Error != nil {
		return nil, resp.Error
	} else {
		return &template, nil
	}
}

func FirstTemplate(query string, params ...interface{}) *Template {
	templates := TemplatesWhere(query, params...)
	if len(templates) == 0 {
		return nil
	}
	return &templates[0]
}

func FirstTemplateUnscoped(query string, params ...interface{}) *Template {
	templates := AllTemplatesWhere(query, params...)
	if len(templates) == 0 {
		return nil
	}
	return &templates[0]
}

func TemplatesWhere(query string, params ...interface{}) []Template {
	var templates []Template
	db.Where(query, params...).Find(&templates)
	return templates
}

func AllTemplatesWhere(query string, params ...interface{}) []Template {
	var templates []Template
	db.Unscoped().Where(query, params...).Find(&templates)
	return templates
}

func CreateTemplate(template Template) (*Template, error) {
	if !db.NewRecord(template) {
		err := template.Save()
		return &template, err
	}

	resp := db.Create(&template)
	return &template, resp.Error
}

func FindAndDeleteTemplate(id interface{}) error {
	template, err := FindTemplate(id)
	if err != nil {
		return err
	}
	if template.DeletedAt != nil {
		return nil // Record has already been deleted.
	}
	return template.Delete()
}

// Updates a single field
func (template *Template) Update(attrs ...interface{}) error {
	return db.Unscoped().Model(template).Update(attrs...).Error
}

// Updates multiple fields
func (template *Template) UpdateFields(fields map[string]interface{}) error {
	return db.Unscoped().Model(template).Updates(fields).Error
}

func (template *Template) Save() error {
	if db.NewRecord(*template) {
		_, err := CreateTemplate(*template)
		return err
	}

	resp := db.Save(template)
	return resp.Error
}

func (template *Template) Delete() error {
	if db.NewRecord(*template) || template.DeletedAt != nil {
		return nil // Record does not exist or has already been deleted.
	}

	resp := db.Delete(template)
	return resp.Error
}

func (template *Template) Reload() error {
	t, err := FindTemplate(template.ID)
	if err != nil {
		return err
	}
	*template = *t

	return nil
}

//================================================================================
// Hooks
//================================================================================

func (template *Template) BeforeSave() (err error) {
	// Validate Required Params
	if err = template.validateRequiredParams(); err != nil {
		return err
	}

	// Validate Unique Params
	if err = template.validateUniqueParams(); err != nil {
		return err
	}

	// Validate Associations
	if err = template.validateAssociations(); err != nil {
		return err
	}

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

func (template *Template) AfterFind() (err error) {
	// Unmarshal JSON
	if err = template.unmarshalJSON(); err != nil {
		return err
	}

	// Deserialize Configurations
	if err = template.deserializeConfigurations(); err != nil {
		return err
	}

	return nil
}

//================================================================================
// Associations
//================================================================================

// --- Validate Associations ---

func (template *Template) validateAssociations() (err error) {
	// Validate Database Type
	if err = template.validateDatabaseType(); err != nil {
		return err
	}

	return nil
}

func (template *Template) validateDatabaseType() error {
	if template.DatabaseTypeID != EmptyUUID { // Try fetching by ID
		databaseType := FirstDatabaseType("id = ? AND domain_id = ?", template.DatabaseTypeID, template.DomainID)

		if databaseType == nil {
			errMsg := fmt.Sprintf("No database type found with ID: %v", template.DatabaseTypeID)
			return errors.New(errMsg)
		}

		if databaseType.ComingSoon {
			return errors.New("database types with the flag 'Coming soon' cannot have a template")
		}

		template.DatabaseTypeName = databaseType.Name
	} else if template.DatabaseTypeName != "" { // Try fetching by name
		databaseType := FirstDatabaseType("name = ? AND domain_id = ?", template.DatabaseTypeName, template.DomainID)

		if databaseType == nil {
			errMsg := fmt.Sprintf("No database type found with name: %v", template.DatabaseTypeName)
			return errors.New(errMsg)
		}

		if databaseType.ComingSoon {
			return errors.New("database types with the flag 'Coming soon' cannot have a template")
		}

		template.DatabaseTypeID = databaseType.ID
	}

	return nil
}

//================================================================================
// Model Validations
//================================================================================

func (template *Template) validateRequiredParams() error {
	var missingParams []string

	if template.Name == "" {
		missingParams = append(missingParams, "name")
	}
	if template.DatabaseTypeID == EmptyUUID && template.DatabaseTypeName == "" {
		missingParams = append(missingParams, "database_type_id or database_type")
	}
	if _, cpuSet := template.Resources["cpu"]; !cpuSet {
		missingParams = append(missingParams, "resources.cpu")
	}
	if _, memSet := template.Resources["memory"]; !memSet {
		missingParams = append(missingParams, "resources.memory")
	}
	if _, diskSet := template.Resources["disk"]; !diskSet {
		missingParams = append(missingParams, "resources.disk")
	}
	if template.DomainID == EmptyUUID {
		missingParams = append(missingParams, "domain_id")
	}

	if len(missingParams) > 0 {
		errMsg := fmt.Sprintf("Missing required param(s): %v", strings.Join(missingParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}

func (template *Template) validateUniqueParams() error {
	invalid := false

	// Assign database type name if missing
	if template.DatabaseTypeName == "" {
		if err := template.validateDatabaseType(); err != nil {
			return err
		}
	}

	expectedResults := 0
	if template.ID != EmptyUUID {
		expectedResults = 1
	}

	results := TemplatesWhere("name = ? AND database_type = ? AND domain_id = ?", template.Name, template.DatabaseTypeName, template.DomainID)
	if len(results) > expectedResults {
		invalid = true
	}

	if invalid {
		errMsg := fmt.Sprintf("Template with name %v already exists for %v.", template.Name, template.DatabaseTypeName)
		return errors.New(errMsg)
	}

	return nil
}

//================================================================================
// Model Tranformation Helpers
//================================================================================

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

func (template *Template) unmarshalJSON() (err error) {
	// Resources JSON
	err = json.Unmarshal(template.ResourcesJSON.RawMessage, &template.Resources)
	if err != nil {
		return err
	}

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

func (template *Template) deserializeConfigurations() (err error) {
	configsArray := make([]Configuration, 0)

	if len(template.ConfigurationsString) > 1 {
		configs := strings.Split(template.ConfigurationsString, ";;")

		for _, c := range configs {
			config := Configuration{}
			err = json.Unmarshal([]byte(c), &config)
			if err != nil {
				return err
			}
			configsArray = append(configsArray, config)
		}
	}

	template.Configurations = configsArray

	return nil
}
