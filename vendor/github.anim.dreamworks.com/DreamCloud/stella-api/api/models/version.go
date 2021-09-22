package models

import (
	"errors"
	"fmt"
	"strings"

	"github.anim.dreamworks.com/golang/logging"
	"github.com/google/uuid"
)

type Version struct {
	Model
	Name             string    `json:"name"`
	DeploymentCount  int       `json:"deployment_count" gorm:"-"` // Deployment Count is cacluated by the API
	DatabaseTypeID   uuid.UUID `json:"database_type_id"`
	DatabaseTypeName string    `json:"database_type" gorm:"column:database_type"`

	DomainID uuid.UUID `json:"domain_id"`
}

func NewVersion() Version {
	return Version{}
}

func AllVersions() []Version {
	var versions []Version
	db.Unscoped().Find(&versions)
	return versions
}

func VersionsCount(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
) int {
	var count int
	var versions []Version

	sDB := dbWithScope(archived, domainIDs, nil)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("database_type LIKE ?", searchParam)
	}

	sDB.Find(&versions).Count(&count)
	return count
}

func PaginatedVersions(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
	page int,
	perPage int,
	orderBy string,
) []Version {
	var versions []Version
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

	sDB.Find(&versions)
	return versions
}

func FindVersion(id interface{}) (*Version, error) {
	version := NewVersion()
	resp := Find(&version, id)

	if resp.Error != nil {
		return nil, resp.Error
	} else {
		return &version, nil
	}
}

func FirstVersion(query string, params ...interface{}) *Version {
	versions := VersionsWhere(query, params...)
	if len(versions) == 0 {
		return nil
	}
	return &versions[0]
}

func FirstVersionUnscoped(query string, params ...interface{}) *Version {
	versions := AllVersionsWhere(query, params...)
	if len(versions) == 0 {
		return nil
	}
	return &versions[0]
}

func VersionsWhere(query string, params ...interface{}) []Version {
	var versions []Version
	db.Where(query, params...).Find(&versions)
	return versions
}

func AllVersionsWhere(query string, params ...interface{}) []Version {
	var versions []Version
	db.Unscoped().Where(query, params...).Find(&versions)
	return versions
}

func CreateVersion(version Version) (*Version, error) {
	if !db.NewRecord(version) {
		err := version.Save()
		return &version, err
	}

	resp := db.Create(&version)
	return &version, resp.Error
}

func FindAndDeleteVersion(id interface{}) error {
	version, err := FindVersion(id)
	if err != nil {
		return err
	}
	if version.DeletedAt != nil {
		return nil // Record has already been deleted.
	}
	return version.Delete()
}

// Updates a single field
func (version *Version) Update(attrs ...interface{}) error {
	return db.Unscoped().Model(version).Update(attrs...).Error
}

// Updates multiple fields
func (version *Version) UpdateFields(fields map[string]interface{}) error {
	return db.Unscoped().Model(version).Updates(fields).Error
}

func (version *Version) Save() error {
	if db.NewRecord(*version) {
		_, err := CreateVersion(*version)
		return err
	}

	resp := db.Save(version)
	return resp.Error
}

func (version *Version) Delete() error {
	if db.NewRecord(*version) || version.DeletedAt != nil {
		return nil // Record does not exist. Nothing to do.
	}

	resp := db.Delete(version)
	return resp.Error
}

func (version *Version) Reload() error {
	v, err := FindVersion(version.ID)
	if err != nil {
		return err
	}
	*version = *v

	return nil
}

//================================================================================
// Helper Functions
//================================================================================

func (version *Version) SetDeploymentCount() {
	version.DeploymentCount = DeploymentsCountWhere("version_id = ?", version.ID)
}

//================================================================================
// Hooks
//================================================================================

func (version *Version) BeforeSave() (err error) {
	// Validate Required Params
	if err = version.validateRequiredParams(); err != nil {
		return err
	}

	// Validate Unique Params
	if err = version.validateUniqueParams(); err != nil {
		return err
	}

	// Validate Associations
	if err = version.validateAssociations(); err != nil {
		return err
	}

	return nil
}

func (version *Version) AfterUpdate() error {
	return version.updateAssociations()
}

//================================================================================
// Associations
//================================================================================

// --- Validate Associations ---

func (version *Version) validateAssociations() (err error) {
	// Validate Database Type
	if err = version.validateDatabaseType(); err != nil {
		return err
	}

	return nil
}

func (version *Version) validateDatabaseType() error {
	if version.DatabaseTypeID != EmptyUUID { // Try fetching by ID
		databaseType := FirstDatabaseType("id = ? AND domain_id = ?", version.DatabaseTypeID, version.DomainID)

		if databaseType == nil {
			errMsg := fmt.Sprintf("No database type found with ID: %v", version.DatabaseTypeID)
			return errors.New(errMsg)
		}

		version.DatabaseTypeName = databaseType.Name
	} else if version.DatabaseTypeName != "" { // Try fetching by name
		databaseType := FirstDatabaseType("name = ? AND domain_id = ?", version.DatabaseTypeName, version.DomainID)

		if databaseType == nil {
			errMsg := fmt.Sprintf("No database type found with name: %v", version.DatabaseTypeName)
			return errors.New(errMsg)
		}

		version.DatabaseTypeID = databaseType.ID
	}

	return nil
}

// --- Update Associations ---

func (version *Version) updateAssociations() error {
	q := "version_id = ? AND version != ?"

	// Update Deployments
	deployments, err := DeploymentsWhere(q, version.ID, version.Name)
	if err != nil {
		return err
	}
	for _, d := range deployments {
		d.VersionName = version.Name
		if err := d.Save(); err != nil {
			logging.Errorf("Updating deployment %v. %v", d, err)
		}
	}

	// Update Images
	images, err := ImagesWhere(q, version.ID, version.Name)
	if err != nil {
		return err
	}
	for _, i := range images {
		i.VersionName = version.Name
		if err := i.Save(); err != nil {
			logging.Errorf("Updating image %v. %v", i, err)
		}
	}

	return nil
}

//================================================================================
// Model Validations
//================================================================================

func (version *Version) validateRequiredParams() error {
	var missingParams []string

	if version.Name == "" {
		missingParams = append(missingParams, "name")
	}
	if version.DatabaseTypeID == EmptyUUID && version.DatabaseTypeName == "" {
		missingParams = append(missingParams, "database_type_id or database_type")
	}
	if version.DomainID == EmptyUUID {
		missingParams = append(missingParams, "domain_id")
	}

	if len(missingParams) > 0 {
		errMsg := fmt.Sprintf("Missing required param(s): %v", strings.Join(missingParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}

func (version *Version) validateUniqueParams() error {
	invalid := false

	// Assign database type name if missing
	if version.DatabaseTypeName == "" {
		if err := version.validateDatabaseType(); err != nil {
			return err
		}
	}

	expectedResults := 0
	if version.ID != EmptyUUID {
		expectedResults = 1
	}

	results := VersionsWhere("name = ? AND database_type = ? AND domain_id = ?", version.Name, version.DatabaseTypeName, version.DomainID)
	if len(results) > expectedResults {
		invalid = true
	}

	if invalid {
		errMsg := fmt.Sprintf("Version %v of %v already exists.", version.Name, version.DatabaseTypeName)
		return errors.New(errMsg)
	}

	return nil
}
