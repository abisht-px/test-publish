package models

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"

	"github.anim.dreamworks.com/golang/logging"

	"github.anim.dreamworks.com/DreamCloud/stella-api/utils"
)

type DatabaseType struct {
	Model
	Name                 string    `json:"name"`
	ShortName            string    `json:"short_name"`
	CanRegisterInVault   *bool     `json:"can_register_in_vault"`
	HasIncrementalBackup *bool     `json:"has_incremental_backup"`
	HasFullBackup        *bool     `json:"has_full_backup"`
	DeploymentCount      int       `json:"deployment_count" gorm:"-"` // Deployment Count is calculated by the API
	Versions             []Version `json:"-" gorm:"-"`

	DomainID   uuid.UUID `json:"domain_id"`
	ComingSoon bool      `json:"coming_soon"`
}

func NewDatabaseType() DatabaseType {
	return DatabaseType{}
}

func AllDatabaseTypes() []DatabaseType {
	var databaseTypes []DatabaseType
	db.Unscoped().Find(&databaseTypes)
	return databaseTypes
}

func DatabaseTypesCount(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
) int {
	var count int
	var databaseTypes []DatabaseType

	sDB := dbWithScope(archived, domainIDs, nil).Where(queryParams)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("name LIKE ?", searchParam)
	}

	sDB.Find(&databaseTypes).Count(&count)
	return count
}

func PaginatedDatabaseTypes(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
	page int,
	perPage int,
	orderBy string,
) []DatabaseType {
	var databaseTypes []DatabaseType
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

	sDB.Find(&databaseTypes)
	return databaseTypes
}

func FindDatabaseType(id interface{}) (*DatabaseType, error) {
	databaseType := NewDatabaseType()
	resp := Find(&databaseType, id)

	if resp.Error != nil {
		return nil, resp.Error
	} else {
		databaseType.loadAssociations()
		return &databaseType, nil
	}
}

func FirstDatabaseType(query string, params ...interface{}) *DatabaseType {
	// TODO PDS-325: Refactor to use db.First() instead
	databaseTypes := DatabaseTypesWhere(query, params...)
	if len(databaseTypes) == 0 {
		return nil
	}
	return &databaseTypes[0]
}

func FirstDatabaseTypeUnscoped(query string, params ...interface{}) *DatabaseType {
	// TODO PDS-325: Refactor to use db.First() instead
	databaseTypes := AllDatabaseTypesWhere(query, params...)
	if len(databaseTypes) == 0 {
		return nil
	}
	return &databaseTypes[0]
}

func DatabaseTypesWhere(query string, params ...interface{}) []DatabaseType {
	var databaseTypes []DatabaseType
	db.Where(query, params...).Find(&databaseTypes)
	loadDatabaseTypeAssociations(&databaseTypes)
	return databaseTypes
}

func AllDatabaseTypesWhere(query string, params ...interface{}) []DatabaseType {
	var databaseTypes []DatabaseType
	db.Unscoped().Where(query, params...).Find(&databaseTypes)
	loadDatabaseTypeAssociations(&databaseTypes)
	return databaseTypes
}

func CreateDatabaseType(databaseType DatabaseType) (*DatabaseType, error) {
	if !db.NewRecord(databaseType) {
		err := databaseType.Save()
		return &databaseType, err
	}

	resp := db.Create(&databaseType)
	return &databaseType, resp.Error
}

func FindAndDeleteDatabaseType(id interface{}) error {
	databaseType, err := FindDatabaseType(id)
	if err != nil {
		return err
	}
	if databaseType.DeletedAt != nil {
		return nil // Record has already been deleted.
	}
	return databaseType.Delete()
}

// Updates a single field
func (databaseType *DatabaseType) Update(attrs ...interface{}) error {
	return db.Unscoped().Model(databaseType).Update(attrs...).Error
}

// Updates multiple fields
func (databaseType *DatabaseType) UpdateFields(fields map[string]interface{}) error {
	return db.Unscoped().Model(databaseType).Updates(fields).Error
}

func (databaseType *DatabaseType) Save() error {
	if db.NewRecord(*databaseType) {
		_, err := CreateDatabaseType(*databaseType)
		return err
	}

	resp := db.Save(databaseType)
	return resp.Error
}

func (databaseType *DatabaseType) Delete() error {
	if db.NewRecord(*databaseType) || databaseType.DeletedAt != nil {
		return nil // Record does not exist or has already been deleted.
	}

	resp := db.Delete(databaseType)
	return resp.Error
}

func (databaseType *DatabaseType) Reload() error {
	dt, err := FindDatabaseType(databaseType.ID)
	if err != nil {
		return err
	}
	*databaseType = *dt

	return nil
}

//================================================================================
// Helper Functions
//================================================================================

func (databaseType *DatabaseType) SetDeploymentCount() {
	databaseType.DeploymentCount = DeploymentsCountWhere("database_type_id = ?", databaseType.ID)
}

func generateDatabaseTypeShortName(fullName string) string {
	length := math.Min(float64(len(fullName)), 3)
	shortName := fullName[0:int(length)]
	return strings.ToLower(shortName)
}

//================================================================================
// Hooks
//================================================================================
func (databaseType *DatabaseType) BeforeCreate(scope *gorm.Scope) (err error) {
	// Call base Model BeforeCreate hook
	if err = databaseType.Model.BeforeCreate(scope); err != nil {
		return err
	}

	// Set boolean config defaults
	if databaseType.CanRegisterInVault == nil {
		databaseType.CanRegisterInVault = utils.NewFalsePtr()
	}
	if databaseType.HasFullBackup == nil {
		databaseType.HasFullBackup = utils.NewFalsePtr()
	}
	if databaseType.HasIncrementalBackup == nil {
		databaseType.HasIncrementalBackup = utils.NewFalsePtr()
	}

	return nil
}

func (databaseType *DatabaseType) BeforeSave() (err error) {
	// Validate Required Params
	if err = databaseType.validateRequiredParams(); err != nil {
		return err
	}

	// Validate Unique Params
	if err = databaseType.validateUniqueParams(); err != nil {
		return err
	}

	return nil
}

func (databaseType *DatabaseType) AfterUpdate() error {
	return databaseType.updateAssociations()
}

//================================================================================
// Associations
//================================================================================

// --- Loading Associations ---

func loadDatabaseTypeAssociations(databaseTypes *[]DatabaseType) {
	for idx := range *databaseTypes {
		(*databaseTypes)[idx].loadAssociations()
	}
}

func (databaseType *DatabaseType) loadAssociations() {
	databaseType.loadVersions()
}

func (databaseType *DatabaseType) loadVersions() {
	var versions []Version
	db.Model(databaseType).Related(&versions)
	databaseType.Versions = versions
}

// --- Update Associations ---

func (databaseType *DatabaseType) updateAssociations() error {
	query := "database_type_id = ? AND database_type != ?"

	// Update Deployments
	deployments, err := DeploymentsWhere(query, databaseType.ID, databaseType.Name)
	if err != nil {
		return err
	}
	for _, d := range deployments {
		d.DatabaseTypeName = databaseType.Name
		if err := d.Save(); err != nil {
			logging.Errorf("Updating deployment %v. %v", d, err)
		}
	}

	// Update Images
	images, err := ImagesWhere(query, databaseType.ID, databaseType.Name)
	if err != nil {
		return err
	}
	for _, i := range images {
		i.DatabaseTypeName = databaseType.Name
		if err := i.Save(); err != nil {
			logging.Errorf("Updating image %v. %v", i, err)
		}
	}

	// Update Versions
	versions := VersionsWhere(query, databaseType.ID, databaseType.Name)
	for _, v := range versions {
		v.DatabaseTypeName = databaseType.Name
		if err := v.Save(); err != nil {
			logging.Errorf("Updating version %v. %v", v, err)
		}
	}

	// Update Templates
	templates := TemplatesWhere(query, databaseType.ID, databaseType.Name)
	for _, t := range templates {
		t.DatabaseTypeName = databaseType.Name
		if err := t.Save(); err != nil {
			logging.Errorf("Updating template %v. %v", t, err)
		}
	}

	return nil
}

//================================================================================
// Model Validations
//================================================================================

func (databaseType *DatabaseType) validateRequiredParams() error {
	var missingParams []string

	if databaseType.Name == "" {
		missingParams = append(missingParams, "name")
	}
	if databaseType.ShortName == "" {
		missingParams = append(missingParams, "short_name")
	}
	if databaseType.DomainID == EmptyUUID {
		missingParams = append(missingParams, "domain_id")
	}

	if len(missingParams) > 0 {
		errMsg := fmt.Sprintf("Missing required param(s): %v", strings.Join(missingParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}

func (databaseType *DatabaseType) validateUniqueParams() error {
	var invalidParams []string

	expectedResults := 0
	if databaseType.ID != EmptyUUID {
		expectedResults = 1
	}

	results := DatabaseTypesWhere("name = ? AND domain_id = ?", databaseType.Name, databaseType.DomainID)
	if len(results) > expectedResults {
		invalidParams = append(invalidParams, "name")
	}

	if len(invalidParams) > 0 {
		errMsg := fmt.Sprintf("Param(s) must be unique: %v", strings.Join(invalidParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}
