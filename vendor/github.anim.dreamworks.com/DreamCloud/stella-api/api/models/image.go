package models

import (
	"errors"
	"fmt"
	"strings"

	"github.anim.dreamworks.com/golang/logging"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

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

	DatabaseType *DatabaseType `json:"-" gorm:"-"`
	Version      *Version      `json:"-" gorm:"-"`
}

func NewImage() Image {
	return Image{}
}

func AllImages() []Image {
	var images []Image
	db.Unscoped().Find(&images)
	return images
}

func ImagesCount(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
) int {
	var count int
	var images []Image

	sDB := dbWithScope(archived, domainIDs, nil)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("name LIKE ?", searchParam)
	}

	sDB.Find(&images).Count(&count)
	return count
}

func PaginatedImages(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
	page int,
	perPage int,
	orderBy string,
) []Image {
	var images []Image
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

	sDB.Find(&images)
	return images
}

func FindImage(id interface{}) (*Image, error) {
	image := NewImage()
	resp := Find(&image, id)

	if resp == nil || resp.Error != nil {
		return nil, resp.Error
	} else {
		err := image.loadAssociations()
		return &image, err
	}
}

func FirstImage(query string, params ...interface{}) (*Image, error) {
	// TODO PDS-325: Refactor to use db.First() instead
	images, err := ImagesWhere(query, params...)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, nil
	}
	return &images[0], nil
}

func FirstImageUnscoped(query string, params ...interface{}) (*Image, error) {
	// TODO PDS-325: Refactor to use db.First() instead
	images, err := AllImagesWhere(query, params...)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, nil
	}
	return &images[0], nil
}

func ImagesWhere(query string, params ...interface{}) ([]Image, error) {
	var images []Image
	db.Where(query, params...).Find(&images)
	err := loadImageAssociations(&images)
	return images, err
}

func AllImagesWhere(query string, params ...interface{}) ([]Image, error) {
	var images []Image
	db.Unscoped().Where(query, params...).Find(&images)
	err := loadImageAssociations(&images)
	return images, err
}

func CreateImage(image Image) (*Image, error) {
	if !db.NewRecord(image) {
		err := image.Save()
		return &image, err
	}

	resp := db.Create(&image)
	return &image, resp.Error
}

func FindAndDeleteImage(id interface{}) error {
	image, err := FindImage(id)
	if err != nil {
		return err
	}

	if image == nil || image.DeletedAt != nil {
		return nil // Record does not exist or has already been deleted.
	}
	if err := image.Delete(); err != nil {
		return err
	}
	return nil
}

// Updates a single field
func (image *Image) Update(attrs ...interface{}) error {
	return db.Unscoped().Model(image).Update(attrs...).Error
}

// Updates multiple fields
func (image *Image) UpdateFields(fields map[string]interface{}) error {
	return db.Unscoped().Model(image).Updates(fields).Error
}

func (image *Image) Save() error {
	if db.NewRecord(*image) {
		_, err := CreateImage(*image)
		return err
	}

	resp := db.Save(image)
	return resp.Error
}

func (image *Image) Delete() error {
	if db.NewRecord(*image) || image.DeletedAt != nil {
		return nil // Record does not exist or has already been deleted.
	}

	resp := db.Delete(image)
	return resp.Error
}

func (image *Image) Reload() error {
	i, err := FindImage(image.ID)
	if err != nil {
		return err
	}

	if i == nil {
		return fmt.Errorf("Image with ID %v not found. Could not reload.", image.ID)
	}
	*image = *i

	return nil
}

//================================================================================
// Helper Functions
//================================================================================

func (image *Image) FriendlyName() string {
	return fmt.Sprintf("%v %v", image.Name, image.Tag)
}

func (image *Image) LongName() string {
	return fmt.Sprintf("%v/%v/%v", image.Registry, image.Namespace, image.Name)
}

func (image *Image) FullName() string {
	return fmt.Sprintf("%v/%v/%v:%v", image.Registry, image.Namespace, image.Name, image.Tag)
}

func (image *Image) SetDeploymentCount() {
	image.DeploymentCount = DeploymentsCountWhere("image_id = ?", image.ID)
}

//================================================================================
// Hooks
//================================================================================

func (image *Image) BeforeSave() (err error) {
	// Validate Required Params
	if err = image.validateRequiredParams(); err != nil {
		return err
	}

	// Validate Unique Params
	if err = image.validateUniqueParams(); err != nil {
		return err
	}

	// Validate Associations
	if err = image.validateAssociations(); err != nil {
		return err
	}

	return nil
}

func (image *Image) BeforeCreate(scope *gorm.Scope) (err error) {
	// Call the parent Model's BeforeCreate hook
	if err = image.Model.BeforeCreate(scope); err != nil {
		return err
	}

	// Try to create the Database Type if it does not already exist
	if image.DatabaseTypeName != "" {
		typeQuery := "name = ? AND domain_id = ?"
		dbType := FirstDatabaseType(typeQuery, image.DatabaseTypeName, image.DomainID)
		if dbType == nil {
			dbType = &DatabaseType{
				Name:      image.DatabaseTypeName,
				ShortName: generateDatabaseTypeShortName(image.DatabaseTypeName),
				DomainID:  image.DomainID,
			}
			dbType, err = CreateDatabaseType(*dbType)
			if err != nil {
				return err
			}
			image.DatabaseType = dbType
			image.DatabaseTypeID = dbType.ID
			image.DatabaseTypeName = dbType.Name
		}
	}

	// Try to create the Version if it does not already exist
	if image.VersionName != "" {
		versionQuery := "name = ? AND (database_type_id = ? OR database_type = ?) AND domain_id = ?"
		version := FirstVersion(versionQuery, image.VersionName, image.DatabaseTypeID, image.DatabaseTypeName, image.DomainID)
		if version == nil {
			version = &Version{
				DatabaseTypeID: image.DatabaseTypeID,
				Name:           image.VersionName,
				DomainID:       image.DomainID,
			}
			version, err = CreateVersion(*version)
			if err != nil {
				return err
			}
			image.Version = version
			image.VersionID = version.ID
			image.VersionName = version.Name
		}
	}

	return nil
}

func (image *Image) AfterUpdate() error {
	return image.updateAssociations()
}

//================================================================================
// Associations
//================================================================================

// --- Loading Associations ---

func loadImageAssociations(images *[]Image) error {
	for idx := range *images {
		if err := (*images)[idx].loadAssociations(); err != nil {
			return err
		}
	}
	return nil
}

func (image *Image) loadAssociations() error {
	if err := image.loadDatabaseType(); err != nil {
		return err
	}
	if err := image.loadVersion(); err != nil {
		return err
	}
	return nil
}

func (image *Image) loadDatabaseType() error {
	dbType, err := FindDatabaseType(image.DatabaseTypeID)
	image.DatabaseType = dbType
	return err
}

func (image *Image) loadVersion() error {
	version, err := FindVersion(image.VersionID)
	image.Version = version
	return err
}

// --- Validating Associations

func (image *Image) validateAssociations() (err error) {
	// Validate Database Type
	if err = image.validateDatabaseType(); err != nil {
		return err
	}

	// Validate Version
	if err = image.validateVersion(); err != nil {
		return err
	}

	return nil
}

func (image *Image) validateDatabaseType() error {
	if image.DatabaseTypeID != EmptyUUID { // Try fetching by ID
		databaseType := FirstDatabaseType("id = ? AND domain_id = ?", image.DatabaseTypeID, image.DomainID)

		if databaseType == nil {
			errMsg := fmt.Sprintf("No database type found with ID: %v", image.DatabaseTypeID)
			return errors.New(errMsg)
		}

		image.DatabaseType = databaseType
		image.DatabaseTypeName = databaseType.Name
	} else if image.DatabaseTypeName != "" { // Try fetching by name
		databaseType := FirstDatabaseType("name = ? AND domain_id = ?", image.DatabaseTypeName, image.DomainID)

		if databaseType != nil {
			image.DatabaseType = databaseType
			image.DatabaseTypeID = databaseType.ID
		}

		// If not found, we will create one in the BeforeCreate hook
	}

	return nil
}

func (image *Image) validateVersion() error {
	if image.VersionID != EmptyUUID { // Try fetching by ID
		query := "database_type_id = ? AND id = ? AND domain_id = ?"
		version := FirstVersion(query, image.DatabaseTypeID, image.VersionID, image.DomainID)

		if version == nil {
			errMsg := fmt.Sprintf("Version with ID %v not supported for %v.", image.VersionID, image.DatabaseTypeName)
			return errors.New(errMsg)
		}

		image.Version = version
		image.VersionName = version.Name
	} else if image.VersionName != "" { // Try fetching by name
		query := "database_type_id = ? AND name = ? AND domain_id = ?"
		version := FirstVersion(query, image.DatabaseTypeID, image.VersionName, image.DomainID)

		if version != nil {
			image.Version = version
			image.VersionID = version.ID
		}

		// If not found, we will create one in the BeforeCreate hook
	}

	return nil
}

// --- Update Associations ---

func (image *Image) updateAssociations() error {
	query := "image_id = ? AND image != ?"

	// Update Deployments
	deployments, err := DeploymentsWhere(query, image.ID, image.FriendlyName())
	if err != nil {
		return err
	}

	for _, d := range deployments {
		d.ImageName = image.FriendlyName()
		if err := d.Save(); err != nil {
			logging.Errorf("Upating deployment %v. %v", d, err)
		}
	}
	return nil
}

//================================================================================
// Model Validations
//================================================================================

func (image *Image) validateRequiredParams() error {
	var missingParams []string

	if image.Registry == "" {
		missingParams = append(missingParams, "registry")
	}
	if image.Namespace == "" {
		missingParams = append(missingParams, "namespace")
	}
	if image.Name == "" {
		missingParams = append(missingParams, "name")
	}
	if image.Build == "" {
		missingParams = append(missingParams, "build")
	}
	if image.Environments == "" {
		missingParams = append(missingParams, "environments")
	}
	if image.DatabaseTypeID == EmptyUUID && image.DatabaseTypeName == "" {
		missingParams = append(missingParams, "database_type_id or database_type")
	}
	if image.VersionID == EmptyUUID && image.VersionName == "" {
		missingParams = append(missingParams, "version_id or version")
	}
	if image.DomainID == EmptyUUID {
		missingParams = append(missingParams, "domain_id")
	}

	if len(missingParams) > 0 {
		errMsg := fmt.Sprintf("Missing required param(s): %v", strings.Join(missingParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}

func (image *Image) validateUniqueParams() error {
	invalid := false

	expectedResults := 0
	if image.ID != EmptyUUID {
		expectedResults = 1
	}

	query := "registry = ? AND namespace = ? AND name = ? AND tag = ? AND deleted_at IS NULL"
	results, err := ImagesWhere(query, image.Registry, image.Namespace, image.Name, image.Tag)
	if err != nil {
		return err
	}
	if len(results) > expectedResults {
		invalid = true
	}

	if invalid {
		errStr := "Registry, namespace, name, and tag must be a unique combination. %v %v already exists."
		errMsg := fmt.Sprintf(errStr, image.LongName(), image.Tag)
		return errors.New(errMsg)
	}

	return nil
}
