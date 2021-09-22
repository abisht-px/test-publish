package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type TargetCluster struct {
	Model
	Name      string      `json:"name"`
	Token     string      `json:"token"`
	APIServer string      `json:"api_server"`
	Images    []Image     `json:"-" gorm:"many2many:target_cluster_images;foreignKey:ID;joinForeignKey:ID"`
	ImageIDs  []uuid.UUID `json:"image_ids" gorm:"-"`

	DomainID uuid.UUID `json:"domain_id"`
}

func NewTargetCluster() TargetCluster {
	return TargetCluster{}
}

func AllTargetClusters() []TargetCluster {
	var targetClusters []TargetCluster
	db.Unscoped().Find(&targetClusters)
	return targetClusters
}

func TargetClustersCount(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
) int {
	var count int
	var targetClusters []TargetCluster

	sDB := dbWithScope(archived, domainIDs, nil).Where(queryParams)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("name LIKE ?", searchParam)
	}

	sDB.Find(&targetClusters).Count(&count)
	return count
}

func PaginatedTargetClusters(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
	page int,
	perPage int,
	orderBy string,
) []TargetCluster {
	var targetClusters []TargetCluster
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

	sDB.Find(&targetClusters)
	return targetClusters
}

func FindTargetCluster(id interface{}) (*TargetCluster, error) {
	targetCluster := NewTargetCluster()
	resp := Find(&targetCluster, id)

	if resp.Error != nil {
		return nil, resp.Error
	} else {
		return &targetCluster, nil
	}
}

func FirstTargetCluster(query string, params ...interface{}) *TargetCluster {
	// TODO PDS-325: Refactor to use db.First() instead
	targetClusters := TargetClustersWhere(query, params...)
	if len(targetClusters) == 0 {
		return nil
	}
	return &targetClusters[0]
}

func FirstTargetClusterUnscoped(query string, params ...interface{}) *TargetCluster {
	// TODO PDS-325: Refactor to use db.First() instead
	targetClusters := AllTargetClustersWhere(query, params...)
	if len(targetClusters) == 0 {
		return nil
	}
	return &targetClusters[0]
}

func TargetClustersWhere(query string, params ...interface{}) []TargetCluster {
	var targetClusters []TargetCluster
	db.Where(query, params...).Find(&targetClusters)
	return targetClusters
}

func AllTargetClustersWhere(query string, params ...interface{}) []TargetCluster {
	var targetClusters []TargetCluster
	db.Unscoped().Where(query, params...).Find(&targetClusters)
	return targetClusters
}

func CreateTargetCluster(targetCluster TargetCluster) (*TargetCluster, error) {
	if !db.NewRecord(targetCluster) {
		err := targetCluster.Save()
		return &targetCluster, err
	}

	resp := db.Create(&targetCluster)
	return &targetCluster, resp.Error
}

func FindAndDeleteTargetCluster(id interface{}) error {
	targetCluster, err := FindTargetCluster(id)
	if err != nil {
		return err
	}
	if targetCluster.DeletedAt != nil {
		return nil // Record has already been deleted.
	}
	return targetCluster.Delete()
}

// Updates a single field
func (targetCluster *TargetCluster) Update(attrs ...interface{}) error {
	return db.Unscoped().Model(targetCluster).Update(attrs...).Error
}

// Updates multiple fields
func (targetCluster *TargetCluster) UpdateFields(fields map[string]interface{}) error {
	return db.Unscoped().Model(targetCluster).Updates(fields).Error
}

func (targetCluster *TargetCluster) Save() error {
	if db.NewRecord(*targetCluster) {
		_, err := CreateTargetCluster(*targetCluster)
		return err
	}

	resp := db.Save(targetCluster)
	return resp.Error
}

func (targetCluster *TargetCluster) Delete() error {
	if db.NewRecord(*targetCluster) || targetCluster.DeletedAt != nil {
		return nil // Record does not exist or has already been deleted.
	}

	resp := db.Delete(targetCluster)
	return resp.Error
}

func (targetCluster *TargetCluster) Reload() error {
	t, err := FindTargetCluster(targetCluster.ID)
	if err != nil {
		return err
	}
	*targetCluster = *t

	return nil
}

//================================================================================
// Helper Functions
//================================================================================

func (targetCluster *TargetCluster) SetImages() error {
	targetCluster.Images = []Image{}
	for _, imageID := range targetCluster.ImageIDs {
		image, err := FindImage(imageID)
		if err != nil {
			return err
		}
		if image != nil {
			targetCluster.Images = append(targetCluster.Images, *image)
		}
	}
	return nil
}

//================================================================================
// Hooks
//================================================================================

func (targetCluster *TargetCluster) BeforeSave() (err error) {
	// Validate Required Params
	if err = targetCluster.validateRequiredParams(); err != nil {
		return err
	}

	// Validate Unique Params
	if err = targetCluster.validateUniqueParams(); err != nil {
		return err
	}

	return nil
}

func (targetCluster *TargetCluster) BeforeUpdate() (err error) {
	// Update images assocation
	if err = db.Model(&targetCluster).Association("Images").Replace(targetCluster.Images).Error; err != nil {
		return err
	}

	return nil
}

func (targetCluster *TargetCluster) AfterUpdate() (err error) {
	if err = targetCluster.updateAssociations(); err != nil {
		return err
	}

	return nil
}

func (targetCluster *TargetCluster) AfterFind() {
	// Load Images
	var images []Image
	db.Model(&targetCluster).Association("Images").Find(&images)
	targetCluster.Images = images

	// Set ImageIDs
	var ids []uuid.UUID
	for _, image := range images {
		ids = append(ids, image.ID)
	}
	targetCluster.ImageIDs = ids
}

//================================================================================
// Associations
//================================================================================

// --- Update Associations ---

func (targetCluster *TargetCluster) updateAssociations() (err error) {
	query := "target_cluster_id = ? AND target_cluster != ?"

	// Update Deployments
	deployments, err := DeploymentsWhere(query, targetCluster.ID, targetCluster.Name)
	if err != nil {
		return err
	}

	for _, d := range deployments {
		d.TargetClusterName = targetCluster.Name
		if err = d.Save(); err != nil {
			return err
		}
	}

	return nil
}

//================================================================================
// Model Validations
//================================================================================

func (targetCluster *TargetCluster) validateRequiredParams() error {
	var missingParams []string

	if targetCluster.Name == "" {
		missingParams = append(missingParams, "name")
	}
	if targetCluster.DomainID == EmptyUUID {
		missingParams = append(missingParams, "domain_id")
	}

	if len(missingParams) > 0 {
		errMsg := fmt.Sprintf("Missing required param(s): %v", strings.Join(missingParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}

func (targetCluster *TargetCluster) validateUniqueParams() error {
	var invalidParams []string

	expectedResults := 0
	if targetCluster.ID != EmptyUUID {
		expectedResults = 1
	}

	results := TargetClustersWhere("name = ? AND domain_id = ?", targetCluster.Name, targetCluster.DomainID)
	if len(results) > expectedResults {
		invalidParams = append(invalidParams, "name")
	}

	if len(invalidParams) > 0 {
		errMsg := fmt.Sprintf("Param(s) must be unique: %v", strings.Join(invalidParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}
