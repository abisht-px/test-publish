package models

import (
	"errors"
	"fmt"
	"strings"

	"github.anim.dreamworks.com/golang/logging"
	"github.com/google/uuid"
)

type DeploymentGroup struct {
	Model
	Name string `json:"name"`

	DomainID uuid.UUID `json:"domain_id"`
}

func NewDeploymentGroup() DeploymentGroup {
	return DeploymentGroup{}
}

func AllDeploymentGroups() []DeploymentGroup {
	var deploymentGroups []DeploymentGroup
	db.Unscoped().Find(&deploymentGroups)
	return deploymentGroups
}

func DeploymentGroupsCount(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
) int {
	var count int
	var deploymentGroups []DeploymentGroup

	sDB := dbWithScope(archived, domainIDs, nil).Where(queryParams)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("name LIKE ?", searchParam)
	}

	sDB.Find(&deploymentGroups).Count(&count)
	return count
}

func PaginatedDeploymentGroups(
	archived string,
	domainIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
	page int,
	perPage int,
	orderBy string,
) []DeploymentGroup {
	var deploymentGroups []DeploymentGroup
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

	sDB.Find(&deploymentGroups)
	return deploymentGroups
}

func FindDeploymentGroup(id interface{}) (*DeploymentGroup, error) {
	deploymentGroup := NewDeploymentGroup()
	resp := Find(&deploymentGroup, id)

	if resp.Error != nil {
		return nil, resp.Error
	} else {
		return &deploymentGroup, nil
	}
}

func FirstDeploymentGroup(query string, params ...interface{}) *DeploymentGroup {
	// TODO PDS-325: Refactor to use db.First() instead
	deploymentGroups := DeploymentGroupsWhere(query, params...)
	if len(deploymentGroups) == 0 {
		return nil
	}
	return &deploymentGroups[0]
}

func FirstDeploymentGroupUnscoped(query string, params ...interface{}) *DeploymentGroup {
	// TODO PDS-325: Refactor to use db.First() instead
	deploymentGroups := AllDeploymentGroupsWhere(query, params...)
	if len(deploymentGroups) == 0 {
		return nil
	}
	return &deploymentGroups[0]
}

func DeploymentGroupsWhere(query string, params ...interface{}) []DeploymentGroup {
	var deploymentGroups []DeploymentGroup
	db.Where(query, params...).Find(&deploymentGroups)
	return deploymentGroups
}

func AllDeploymentGroupsWhere(query string, params ...interface{}) []DeploymentGroup {
	var deploymentGroups []DeploymentGroup
	db.Unscoped().Where(query, params...).Find(&deploymentGroups)
	return deploymentGroups
}

func CreateDeploymentGroup(deploymentGroup DeploymentGroup) (*DeploymentGroup, error) {
	if !db.NewRecord(deploymentGroup) {
		err := deploymentGroup.Save()
		return &deploymentGroup, err
	}

	resp := db.Create(&deploymentGroup)
	return &deploymentGroup, resp.Error
}

func FindAndDeleteDeploymentGroup(id interface{}) error {
	deploymentGroup, err := FindDeploymentGroup(id)
	if err != nil {
		return err
	}
	if deploymentGroup.DeletedAt != nil {
		return nil // Record has already been deleted.
	}
	return deploymentGroup.Delete()
}

// Updates a single field
func (deploymentGroup *DeploymentGroup) Update(attrs ...interface{}) error {
	return db.Unscoped().Model(deploymentGroup).Update(attrs...).Error
}

// Updates multiple fields
func (deploymentGroup *DeploymentGroup) UpdateFields(fields map[string]interface{}) error {
	return db.Unscoped().Model(deploymentGroup).Updates(fields).Error
}

func (deploymentGroup *DeploymentGroup) Save() error {
	if db.NewRecord(*deploymentGroup) {
		_, err := CreateDeploymentGroup(*deploymentGroup)
		return err
	}

	resp := db.Save(deploymentGroup)
	return resp.Error
}

func (deploymentGroup *DeploymentGroup) Delete() error {
	if db.NewRecord(*deploymentGroup) || deploymentGroup.DeletedAt != nil {
		return nil // Record does not exist or has already been deleted.
	}

	resp := db.Delete(deploymentGroup)
	return resp.Error
}

func (deploymentGroup *DeploymentGroup) Reload() error {
	d, err := FindDeploymentGroup(deploymentGroup.ID)
	if err != nil {
		return err
	}
	if d == nil {
		return fmt.Errorf("DeploymentGroup with ID %v not found. Could not reload.", deploymentGroup.ID)
	}
	*deploymentGroup = *d

	return nil
}

func (deploymentGroup *DeploymentGroup) AfterUpdate() error {
	return deploymentGroup.updateAssociations()
}

//================================================================================
// Hooks
//================================================================================

func (deploymentGroup *DeploymentGroup) BeforeSave() (err error) {
	// Validate Required Params
	if err = deploymentGroup.validateRequiredParams(); err != nil {
		return err
	}

	// Validate Unique Params
	return deploymentGroup.validateUniqueParams()
}

//================================================================================
// Associations
//================================================================================

// --- Update Associations ---

func (deploymentGroup *DeploymentGroup) updateAssociations() error {
	query := "deployment_group_id = ? AND deployment_group != ?"

	// Update Deployments
	deployments, err := DeploymentsWhere(query, deploymentGroup.ID, deploymentGroup.Name)
	if err != nil {
		return err
	}

	for _, d := range deployments {
		d.DeploymentGroupName = deploymentGroup.Name
		if err := d.Save(); err != nil {
			logging.Errorf("Updating deployment %v. %v", d, err)
		}
	}
	return nil
}

//================================================================================
// Model Validations
//================================================================================

func (deploymentGroup *DeploymentGroup) validateRequiredParams() error {
	var missingParams []string

	if deploymentGroup.Name == "" {
		missingParams = append(missingParams, "name")
	}
	if deploymentGroup.DomainID == EmptyUUID {
		missingParams = append(missingParams, "domain_id")
	}

	if len(missingParams) > 0 {
		errMsg := fmt.Sprintf("Missing required param(s): %v", strings.Join(missingParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}

func (deploymentGroup *DeploymentGroup) validateUniqueParams() error {
	var invalidParams []string

	expectedResults := 0
	if deploymentGroup.ID != EmptyUUID {
		expectedResults = 1
	}

	results := DeploymentGroupsWhere("name = ?", deploymentGroup.Name)
	if len(results) > expectedResults {
		invalidParams = append(invalidParams, "name")
	}

	if len(invalidParams) > 0 {
		errMsg := fmt.Sprintf("Param(s) must be unique: %v", strings.Join(invalidParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}
