package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Environment struct {
	Model
	Name             string          `json:"name"`
	TargetClusters   []TargetCluster `json:"-" gorm:"many2many:environment_target_clusters;foreignKey:ID;joinForeignKey:ID"`
	TargetClusterIDs []uuid.UUID     `json:"target_cluster_ids" gorm:"-"`

	DomainID uuid.UUID `json:"domain_id"`
}

func NewEnvironment() Environment {
	return Environment{}
}

func AllEnvironments() []Environment {
	var environments []Environment
	db.Unscoped().Find(&environments)
	return environments
}

func EnvironmentsCount(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
) int {
	var count int
	var environments []Environment

	sDB := dbWithScope(archived, domainIDs, nil).Where(queryParams)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("name LIKE ?", searchParam)
	}

	sDB.Find(&environments).Count(&count)
	return count
}

func PaginatedEnvironments(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
	page int,
	perPage int,
	orderBy string,
) []Environment {
	var environments []Environment
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
		sDB = sDB.Where("name LIKE ?", searchParam)
	}

	sDB.Find(&environments)
	return environments
}

func FindEnvironment(id interface{}) (*Environment, error) {
	environment := NewEnvironment()
	resp := Find(&environment, id)

	if resp.Error != nil {
		return nil, resp.Error
	} else {
		return &environment, nil
	}
}

func FirstEnvironment(query string, params ...interface{}) *Environment {
	// TODO PDS-325: Refactor to use db.First() instead
	environments := EnvironmentsWhere(query, params...)
	if len(environments) == 0 {
		return nil
	}
	return &environments[0]
}

func FirstEnvironmentUnscoped(query string, params ...interface{}) *Environment {
	// TODO PDS-325: Refactor to use db.First() instead
	environments := AllEnvironmentsWhere(query, params...)
	if len(environments) == 0 {
		return nil
	}
	return &environments[0]
}

func EnvironmentsWhere(query string, params ...interface{}) []Environment {
	var environments []Environment
	db.Where(query, params...).Find(&environments)
	return environments
}

func AllEnvironmentsWhere(query string, params ...interface{}) []Environment {
	var environments []Environment
	db.Unscoped().Where(query, params...).Find(&environments)
	return environments
}

func CreateEnvironment(environment Environment) (*Environment, error) {
	if !db.NewRecord(environment) {
		err := environment.Save()
		return &environment, err
	}

	resp := db.Create(&environment)
	return &environment, resp.Error
}

func FindAndDeleteEnvironment(id interface{}) error {
	environment, err := FindEnvironment(id)
	if err != nil {
		return err
	}
	if environment.DeletedAt != nil {
		return nil // Record has already been deleted.
	}
	return environment.Delete()
}

// Updates a single field
func (environment *Environment) Update(attrs ...interface{}) error {
	return db.Unscoped().Model(environment).Update(attrs...).Error
}

// Updates multiple fields
func (environment *Environment) UpdateFields(fields map[string]interface{}) error {
	return db.Unscoped().Model(environment).Updates(fields).Error
}

func (environment *Environment) Save() error {
	if db.NewRecord(*environment) {
		_, err := CreateEnvironment(*environment)
		return err
	}

	resp := db.Save(environment)
	return resp.Error
}

func (environment *Environment) Delete() error {
	if db.NewRecord(*environment) || environment.DeletedAt != nil {
		return nil // Record does not exist or has already been deleted.
	}

	resp := db.Delete(environment)
	return resp.Error
}

func (environment *Environment) Reload() error {
	e, err := FindEnvironment(environment.ID)
	if err != nil {
		return err
	}
	*environment = *e

	return nil
}

//================================================================================
// Helper Functions
//================================================================================

func (environment *Environment) SetTargetClusters() {
	environment.TargetClusters = []TargetCluster{}
	for _, targetClusterID := range environment.TargetClusterIDs {
		targetCluster, err := FindTargetCluster(targetClusterID)
		if err == nil && targetCluster != nil {
			environment.TargetClusters = append(environment.TargetClusters, *targetCluster)
		}
	}
}

//================================================================================
// Hooks
//================================================================================

func (environment *Environment) BeforeSave() (err error) {
	// Validate Required Params
	if err = environment.validateRequiredParams(); err != nil {
		return err
	}

	// Validate Unique Params
	if err = environment.validateUniqueParams(); err != nil {
		return err
	}

	return nil
}

func (environment *Environment) BeforeUpdate() (err error) {
	// Update target cluster associations
	if err = db.Model(&environment).Association("TargetClusters").Replace(environment.TargetClusters).Error; err != nil {
		return err
	}

	return nil
}

func (environment *Environment) AfterUpdate() (err error) {
	if err = environment.updateAssociations(); err != nil {
		return err
	}

	return nil
}

func (environment *Environment) AfterFind() {
	// Load Target Clusters
	var targetClusters []TargetCluster
	db.Model(&environment).Association("TargetClusters").Find(&targetClusters)
	environment.TargetClusters = targetClusters

	// Set Target Cluster IDs
	var ids []uuid.UUID
	for _, cluster := range targetClusters {
		ids = append(ids, cluster.ID)
	}
	environment.TargetClusterIDs = ids
}

//================================================================================
// Associations
//================================================================================

// --- Update Associations ---

func (environment *Environment) updateAssociations() (err error) {
	query := "environment_id = ? AND environment != ?"

	// Update Deployments
	deployments, err := DeploymentsWhere(query, environment.ID, environment.Name)
	if err != nil {
		return err
	}
	for _, d := range deployments {
		d.EnvironmentName = environment.Name
		err = d.Save()
		if err != nil {
			return err
		}
	}

	return nil
}

//================================================================================
// Model Validations
//================================================================================

func (environment *Environment) validateRequiredParams() error {
	var missingParams []string

	if environment.Name == "" {
		missingParams = append(missingParams, "name")
	}
	if environment.DomainID == EmptyUUID {
		missingParams = append(missingParams, "domain_id")
	}

	if len(missingParams) > 0 {
		errMsg := fmt.Sprintf("Missing required param(s): %v", strings.Join(missingParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}

func (environment *Environment) validateUniqueParams() error {
	var invalidParams []string

	expectedResults := 0
	if environment.ID != EmptyUUID {
		expectedResults = 1
	}

	results := EnvironmentsWhere("name = ? AND domain_id = ?", environment.Name, environment.DomainID)
	if len(results) > expectedResults {
		invalidParams = append(invalidParams, "name")
	}

	if len(invalidParams) > 0 {
		errMsg := fmt.Sprintf("Param(s) must be unique: %v", strings.Join(invalidParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}
