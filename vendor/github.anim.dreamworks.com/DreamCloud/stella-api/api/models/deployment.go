package models

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.anim.dreamworks.com/DreamCloud/stella-api/utils"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type Deployment struct {
	Model
	NodeCount                 uint                   `json:"node_count" mapstructure:"node_count"`
	RegisteredInVault         bool                   `json:"registered_in_vault" mapstructure:"-"`       // RegisteredInVault is managed by the API
	DeploymentID              string                 `json:"deployment_id" mapstructure:"deployment_id"` // DeploymentID is auto generated based on the cluster name and build
	ClusterName               string                 `json:"cluster_name" mapstructure:"cluster_name"`
	Build                     string                 `json:"build" mapstructure:"build"`                   // Build will be auto generated (you also have the ability to provide your own - must be unique)
	Origin                    string                 `json:"origin" mapstructure:"-" enum:"stella,legacy"` // Origin defaults to 'stella'
	State                     string                 `json:"state" mapstructure:"-"`                       // State is managed by the API
	Schema                    string                 `json:"schema" mapstructure:"schema"`
	Endpoint                  string                 `json:"endpoint" mapstructure:"-"` // Endpoint is set by the Tekton pipeline on creation
	Service                   string                 `json:"service" mapstructure:"service"`
	ServiceType               string                 `json:"service_type" mapstructure:"service_type"`
	StorageProvider           string                 `json:"storage_provider" mapstructure:"storage_provider" enum:"portworx,trident"`
	FullBackupSchedule        string                 `json:"full_backup_schedule" mapstructure:"-"`
	IncrementalBackupSchedule string                 `json:"incremental_backup_schedule" mapstructure:"-"`
	FullBackupLimit           *uint                  `json:"full_backup_limit" mapstructure:"-"`
	IncrementalBackupLimit    *uint                  `json:"incremental_backup_limit" mapstructure:"-"`
	Configuration             map[string]interface{} `json:"configuration" mapstructure:"configuration" gorm:"-"`
	Resources                 map[string]interface{} `json:"resources" mapstructure:"resources" gorm:"-"`
	ConnectionDetails         map[string]interface{} `json:"connection_details" mapstructure:"connection_details" gorm:"-"`
	LastBackup                *time.Time             `json:"last_backup" mapstructure:"-" binding:"-"` // LastBackup is managed by the API
	DatabaseTypeID            uuid.UUID              `json:"database_type_id" mapstructure:"-"`
	VersionID                 uuid.UUID              `json:"version_id" mapstructure:"-"`
	EnvironmentID             uuid.UUID              `json:"environment_id" mapstructure:"-"`
	ImageID                   *uuid.UUID             `json:"image_id" mapstructure:"-"`
	TargetClusterID           *uuid.UUID             `json:"target_cluster_id" mapstructure:"-"`
	DeploymentGroupID         *uuid.UUID             `json:"deployment_group_id" mapstructure:"-"`
	DatabaseTypeName          string                 `json:"database_type" mapstructure:"database_type" gorm:"column:database_type"`
	VersionName               string                 `json:"version" mapstructure:"version" gorm:"column:version"`
	ImageName                 string                 `json:"image" mapstructure:"image" gorm:"column:image"`
	EnvironmentName           string                 `json:"environment" mapstructure:"environment" gorm:"column:environment"`
	DeploymentGroupName       string                 `json:"deployment_group" mapstructure:"deployment_group" gorm:"column:deployment_group"`
	TargetClusterName         string                 `json:"target_cluster" mapstructure:"target_cluster" gorm:"column:target_cluster"`
	Initialize                string                 `json:"initialize" mapstructure:"initialize"`
	ImagePullSecret           string                 `json:"image_pull_secret" mapstructure:"image_pull_secret"`
	DNSZone                   string                 `json:"dns_zone" mapstructure:"dns_zone"`
	StorageClassProvisioner   string                 `json:"storage_class_provisioner" mapstructure:"storage_class_provisioner"`
	StorageOptions            map[string]string      `json:"storage_options" mapstructure:"storage_options" gorm:"-"`

	DomainID  uuid.UUID `json:"domain_id" mapstructure:"-"`
	ProjectID uuid.UUID `json:"project_id" mapstructure:"-"`
	UserID    uuid.UUID `json:"user_id" mapstructure:"-"`
	UserLogin string    `json:"user_login" mapstructure:"-"`

	DatabaseType    *DatabaseType    `json:"-" mapstructure:"-" gorm:"-"`
	Version         *Version         `json:"-" mapstructure:"-" gorm:"-"`
	Image           *Image           `json:"-" mapstructure:"-" gorm:"-"`
	Environment     *Environment     `json:"-" mapstructure:"-" gorm:"-"`
	DeploymentGroup *DeploymentGroup `json:"-" mapstructure:"-" gorm:"-"`
	TargetCluster   *TargetCluster   `json:"-" mapstructure:"-" gorm:"-"`

	ConfigurationJSON     postgres.Jsonb `json:"-" mapstructure:"-" gorm:"column:configuration"`      // Used for DB
	ResourcesJSON         postgres.Jsonb `json:"-" mapstructure:"-" gorm:"column:resources"`          // Used for DB
	ConnectionDetailsJSON postgres.Jsonb `json:"-" mapstructure:"-" gorm:"column:connection_details"` // Used for DB
	StorageOptionsJSON    postgres.Jsonb `json:"-" mapstructure:"-" gorm:"column:storage_options"`    // Used for DB
}

type DeploymentResources struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Disk   string `json:"disk"`
}

func NewDeployment() Deployment {
	return Deployment{}
}

func AllDeployments() []Deployment {
	var deployments []Deployment
	db.Unscoped().Find(&deployments)
	return deployments
}

func DeploymentsCount(
	archived string,
	domainIDs interface{},
	projectIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
) int {
	var count int
	var deployments []Deployment

	sDB := dbWithScope(archived, domainIDs, projectIDs).Where(queryParams)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("deployment_id LIKE ?", searchParam)
	}

	sDB.Find(&deployments).Count(&count)
	return count
}

func PaginatedDeployments(
	archived string,
	domainIDs interface{},
	projectIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
	page int,
	perPage int,
	orderBy string,
) []Deployment {
	var deployments []Deployment
	offset := (page - 1) * perPage

	if orderBy == "" {
		orderBy = "created_at desc"
	}

	sDB := dbWithScope(archived, domainIDs, projectIDs).Order(orderBy).Offset(offset).Limit(perPage).Where(queryParams)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("deployment_id LIKE ?", searchParam)
	}

	sDB.Find(&deployments)
	return deployments
}

func FindDeployment(id interface{}) (*Deployment, error) {
	deployment := NewDeployment()
	resp := Find(&deployment, id)

	if resp.Error != nil {
		return nil, resp.Error
	} else {
		err := deployment.loadAssociations()
		return &deployment, err
	}
}

func FirstDeployment(query string, params ...interface{}) (*Deployment, error) {
	// TODO PDS-325: Refactor to use db.First() instead
	deployments, err := DeploymentsWhere(query, params...)
	if err != nil {
		return nil, err
	}
	if len(deployments) == 0 {
		return nil, nil
	}
	return &deployments[0], nil
}

func FirstDeploymentUnscoped(query string, params ...interface{}) (*Deployment, error) {
	// TODO PDS-325: Refactor to use db.First() instead
	deployments, err := AllDeploymentsWhere(query, params...)
	if err != nil {
		return nil, err
	}
	if len(deployments) == 0 {
		return nil, nil
	}
	return &deployments[0], nil
}

func DeploymentsWhere(query string, params ...interface{}) ([]Deployment, error) {
	var deployments []Deployment
	db.Where(query, params...).Find(&deployments)
	err := loadDeploymentAssociations(&deployments)
	return deployments, err
}

func AllDeploymentsWhere(query string, params ...interface{}) ([]Deployment, error) {
	var deployments []Deployment
	db.Unscoped().Where(query, params...).Find(&deployments)
	err := loadDeploymentAssociations(&deployments)
	return deployments, err
}

func DeploymentsCountWhere(query string, params ...interface{}) int {
	var count int
	var deployments []Deployment
	db.Where(query, params...).Find(&deployments).Count(&count)
	return count
}

func CreateDeployment(deployment Deployment) (*Deployment, error) {
	if err := deployment.setDefaults(); err != nil {
		return nil, err
	}

	if !db.NewRecord(deployment) {
		err := deployment.Save()
		return &deployment, err
	}

	resp := db.Create(&deployment)
	return &deployment, resp.Error
}

func FindAndDeleteDeployment(id interface{}) error {
	deployment, err := FindDeployment(id)
	if err != nil {
		return err
	}
	if deployment.DeletedAt != nil {
		return nil // Record has already been deleted.
	}
	return deployment.Delete()
}

// Updates a single field
func (deployment *Deployment) Update(attrs ...interface{}) error {
	return db.Unscoped().Model(deployment).Update(attrs...).Error
}

// Updates multiple fields
func (deployment *Deployment) UpdateFields(fields map[string]interface{}) error {
	return db.Unscoped().Model(deployment).Updates(fields).Error
}

func (deployment *Deployment) Save() error {
	if db.NewRecord(*deployment) {
		_, err := CreateDeployment(*deployment)
		return err
	}

	resp := db.Save(deployment)
	return resp.Error
}

func (deployment *Deployment) Delete() error {
	if db.NewRecord(*deployment) || deployment.DeletedAt != nil {
		return nil // Record does not exist or has already been deleted.
	}

	resp := db.Delete(deployment)
	return resp.Error
}

func (deployment *Deployment) Reload() error {
	d, err := FindDeployment(deployment.ID)
	if err != nil {
		return err
	}
	*deployment = *d

	return nil
}

//================================================================================
// Helper Functions
//================================================================================

func (deployment *Deployment) ResourceURI() string {
	return fmt.Sprintf("/api/deployments/%v", deployment.ID)
}

func (deployment *Deployment) SupportsFullBackups() bool {
	val := deployment.DatabaseType.HasFullBackup
	return val != nil && *val
}

func (deployment *Deployment) SupportsIncrementalBackups() bool {
	val := deployment.DatabaseType.HasIncrementalBackup
	return val != nil && *val
}

func (deployment *Deployment) CanRegisterInVault() bool {
	val := deployment.DatabaseType.CanRegisterInVault
	return val != nil && *val
}

//================================================================================
// Hooks
//================================================================================

func (deployment *Deployment) BeforeCreate(scope *gorm.Scope) (err error) {
	// Call Base Model's BeforeCreate()
	if err = deployment.Model.BeforeCreate(scope); err != nil {
		return err
	}

	return nil
}

func (deployment *Deployment) BeforeSave() (err error) {
	// Validate Required Params
	if err = deployment.validateRequiredParams(); err != nil {
		return err
	}
	// Validate Associations
	if err = deployment.validateAssociations(); err != nil {
		return err
	}
	// Validate Params
	if err = deployment.validateParams(); err != nil {
		return err
	}
	// Marshal JSON
	if err = deployment.marshalJSON(); err != nil {
		return err
	}

	return nil
}

func (deployment *Deployment) AfterFind() error {
	return deployment.unmarshalJSON()
}

//================================================================================
// Associations
//================================================================================

// --- Loading Associations ---

func loadDeploymentAssociations(deployments *[]Deployment) error {
	for idx := range *deployments {
		if err := (*deployments)[idx].loadAssociations(); err != nil {
			return err
		}
	}
	return nil
}

func (deployment *Deployment) loadAssociations() error {
	if err := deployment.loadDatabaseType(); err != nil {
		return err
	}
	if err := deployment.loadVersion(); err != nil {
		return err
	}
	if err := deployment.loadEnvironment(); err != nil {
		return err
	}

	if deployment.ImageID != nil {
		if err := deployment.loadImage(); err != nil {
			return err
		}
	}
	if deployment.DeploymentGroupID != nil {
		if err := deployment.loadDeploymentGroup(); err != nil {
			return err
		}
	}
	if deployment.TargetClusterID != nil {
		if err := deployment.loadTargetCluster(); err != nil {
			return err
		}
	}
	return nil
}

func (deployment *Deployment) loadDatabaseType() error {
	dt, err := FindDatabaseType(deployment.DatabaseTypeID)
	if err != nil {
		return err
	}
	deployment.DatabaseType = dt
	return nil
}

func (deployment *Deployment) loadVersion() error {
	ver, err := FindVersion(deployment.VersionID)
	if err != nil {
		return err
	}
	deployment.Version = ver
	return nil
}

func (deployment *Deployment) loadEnvironment() error {
	env, err := FindEnvironment(deployment.EnvironmentID)
	if err != nil {
		return err
	}
	deployment.Environment = env
	return nil
}

func (deployment *Deployment) loadImage() error {
	img, err := FindImage(deployment.ImageID)
	deployment.Image = img
	return err
}

func (deployment *Deployment) loadDeploymentGroup() error {
	dp, err := FindDeploymentGroup(deployment.DeploymentGroupID)
	if err != nil {
		return err
	}
	deployment.DeploymentGroup = dp
	return nil
}

func (deployment *Deployment) loadTargetCluster() error {
	tc, err := FindTargetCluster(deployment.TargetClusterID)
	if err != nil {
		return err
	}
	deployment.TargetCluster = tc
	return nil
}

// --- Defaults ---

func (deployment *Deployment) setDefaults() error {
	// Set Deployment Origin
	if deployment.Origin == "" {
		deployment.Origin = "stella"
	}

	if _, ok := deployment.Resources["storage"]; !ok {
		// Database operator expects the resources.storage parameter while
		// dashboard is setting the resources.disk one. (PDS-225)
		deployment.Resources["storage"] = deployment.Resources["disk"]
	}

	// These parameters are required by the new deployment pipeline. Set default values for them
	// so it's possible to deploy databases with UI.
	if deployment.Initialize == "" {
		deployment.Initialize = "Manual"
	}
	if deployment.StorageClassProvisioner == "" {
		deployment.StorageClassProvisioner = "kubernetes.io/portworx-volume"
	}
	if deployment.ImagePullSecret == "" {
		deployment.ImagePullSecret = "pds-docker-registry-credentials"
	}
	if deployment.DNSZone == "" {
		deployment.DNSZone = "pds-dev.io"
	}
	if deployment.StorageOptions == nil {
		deployment.StorageOptions = map[string]string{
			"replicas":      "2",
			"group":         deployment.ClusterName,
			"forceSpread":   "false",
			"reclaimPolicy": "Retain",
		}
	}

	// Generate build (loop to ensure build is unique)
	if deployment.Build == "" {
		build := ""
		for {
			seed := fmt.Sprintf("%v-%v-%v", deployment.ClusterName, deployment.EnvironmentName, time.Now())
			sha := fmt.Sprintf("%x", sha256.Sum256([]byte(seed)))
			build = utils.Substring(sha, 0, 8)

			deployment, err := FirstDeployment("build = ?", build)
			if err != nil {
				return err
			}
			if deployment == nil {
				break // build is unique
			}
		}
		// Set Build
		deployment.Build = build
	} else {
		// Check that provided build is unique
		deployment, err := FirstDeployment("build = ?", deployment.Build)
		if err != nil {
			return err
		}
		if deployment != nil {
			return errors.New("Another deployment exists with the provided build.")
		}
	}

	// Set DeploymentID
	deployment.DeploymentID = fmt.Sprintf("%v-%v", deployment.ClusterName, deployment.Build)

	// PDS-270 investigate whether we need to store the endpoint.
	if deployment.validateTargetCluster() == nil && deployment.validateEnvironment() == nil {
		deployment.Endpoint = fmt.Sprintf("%s/apis/deployments.pds.io/v1/namespaces/%s/databases/%s",
			deployment.TargetCluster.APIServer,
			deployment.EnvironmentName,
			deployment.DeploymentID)
	}

	return nil
}

// --- Validating Associations ---

// NOTE: Ordering matters here
func (deployment *Deployment) validateAssociations() (err error) {
	// Validate Database Type
	if err = deployment.validateDatabaseType(); err != nil {
		return err
	}

	// Validate Version
	if err = deployment.validateVersion(); err != nil {
		return err
	}

	// Validate Environment
	if err = deployment.validateEnvironment(); err != nil {
		return err
	}

	// Validate Deployment Group
	if err = deployment.validateDeploymentGroup(); err != nil {
		return err
	}

	// Validate Target Cluster
	if err = deployment.validateTargetCluster(); err != nil {
		return err
	}

	// Validate Image
	if err = deployment.validateImage(); err != nil {
		return err
	}

	return nil
}

func (deployment *Deployment) validateDatabaseType() error {
	if deployment.DatabaseTypeID != EmptyUUID { // Try fetching by ID
		databaseType := FirstDatabaseType("id = ? AND domain_id = ?", deployment.DatabaseTypeID, deployment.DomainID)

		if databaseType == nil {
			errMsg := fmt.Sprintf("No database type found with ID: %v", deployment.DatabaseTypeID)
			return errors.New(errMsg)
		}

		deployment.DatabaseType = databaseType
		deployment.DatabaseTypeName = databaseType.Name
	} else if deployment.DatabaseTypeName != "" { // Try fetching by name
		databaseType := FirstDatabaseType("name = ? AND domain_id = ?", deployment.DatabaseTypeName, deployment.DomainID)

		if databaseType == nil {
			errMsg := fmt.Sprintf("No database type found with name: %v", deployment.DatabaseTypeName)
			return errors.New(errMsg)
		}

		deployment.DatabaseType = databaseType
		deployment.DatabaseTypeID = databaseType.ID
	}

	return nil
}

func (deployment *Deployment) validateVersion() error {
	if deployment.VersionID != EmptyUUID { // Try fetching by ID
		query := "database_type_id = ? AND id = ? AND domain_id = ?"
		version := FirstVersion(query, deployment.DatabaseTypeID, deployment.VersionID, deployment.DomainID)

		if version == nil {
			errMsg := fmt.Sprintf("Version with ID %v not supported for %v.", deployment.VersionID, deployment.DatabaseTypeName)
			return errors.New(errMsg)
		}

		deployment.Version = version
		deployment.VersionName = version.Name
	} else if deployment.VersionName != "" { // Try fetching by name
		query := "database_type_id = ? AND name = ? AND domain_id = ?"
		version := FirstVersion(query, deployment.DatabaseTypeID, deployment.VersionName, deployment.DomainID)

		if version == nil {
			errMsg := fmt.Sprintf("Version %v of %v is not supported.", deployment.VersionName, deployment.DatabaseTypeName)
			return errors.New(errMsg)
		}

		deployment.Version = version
		deployment.VersionID = version.ID
	}

	return nil
}

func (deployment *Deployment) validateEnvironment() error {
	if deployment.EnvironmentID != EmptyUUID { // Try fetching by ID
		environment := FirstEnvironment("id = ? AND domain_id = ?", deployment.EnvironmentID, deployment.DomainID)

		if environment == nil {
			errMsg := fmt.Sprintf("No environment found with ID: %v", deployment.EnvironmentID)
			return errors.New(errMsg)
		}

		deployment.Environment = environment
		deployment.EnvironmentName = environment.Name
	} else if deployment.EnvironmentName != "" { // Try fetching by name
		environment := FirstEnvironment("name = ? AND domain_id = ?", deployment.EnvironmentName, deployment.DomainID)

		if environment == nil {
			errMsg := fmt.Sprintf("No environment found with name: %v", deployment.EnvironmentName)
			return errors.New(errMsg)
		}

		deployment.Environment = environment
		deployment.EnvironmentID = environment.ID
	}

	return nil
}

func (deployment *Deployment) validateTargetCluster() error {
	if deployment.TargetClusterID == nil && deployment.TargetClusterName == "" {
		return nil
	}

	if deployment.TargetClusterID != nil && *deployment.TargetClusterID != EmptyUUID { // Try fetching by ID
		targetCluster := FirstTargetCluster(
			"id = ? AND domain_id = ?",
			deployment.TargetClusterID,
			deployment.DomainID,
		)

		if targetCluster == nil {
			errMsg := fmt.Sprintf("No target cluster found with ID: %v", deployment.TargetClusterID)
			return errors.New(errMsg)
		}

		deployment.TargetCluster = targetCluster
		deployment.TargetClusterName = targetCluster.Name
	} else if deployment.TargetClusterName != "" { // Try fetching by name
		targetCluster := FirstTargetCluster(
			"name = ? AND domain_id = ?",
			deployment.TargetClusterName,
			deployment.DomainID,
		)

		if targetCluster == nil {
			errMsg := fmt.Sprintf("No target cluster found with name: %v", deployment.TargetClusterName)
			return errors.New(errMsg)
		}

		deployment.TargetCluster = targetCluster
		deployment.TargetClusterID = &targetCluster.ID
	}

	return nil
}

// NOTE: Must be called after validateTargetCluster()
func (deployment *Deployment) validateImage() error {
	if deployment.ImageID == nil {
		return nil
	}

	image, err := FirstImage("id = ? AND domain_id = ?", deployment.ImageID, deployment.DomainID)
	if err != nil {
		return err
	}
	if image == nil {
		errMsg := fmt.Sprintf("No image found with ID: %v", deployment.ImageID)
		return errors.New(errMsg)
	}

	// Check if the image is included as part of the provided target cluster
	if !utils.ContainsUUID(deployment.TargetCluster.ImageIDs, *deployment.ImageID) {
		errMsg := fmt.Sprintf(
			"The image %v is not supported by given target cluster (%v)",
			image.FullName(),
			deployment.TargetClusterName,
		)
		return errors.New(errMsg)
	}

	deployment.Image = image
	deployment.ImageName = image.FriendlyName()

	return nil
}

func (deployment *Deployment) validateDeploymentGroup() error {
	if deployment.DeploymentGroupID == nil && deployment.DeploymentGroupName == "" {
		return nil
	}

	if deployment.DeploymentGroupID != nil && *deployment.DeploymentGroupID != EmptyUUID { // Try fetching by ID
		deploymentGroup := FirstDeploymentGroup(
			"id = ? AND domain_id = ?",
			deployment.DeploymentGroupID,
			deployment.DomainID,
		)

		if deploymentGroup == nil {
			errMsg := fmt.Sprintf("No deployment group found with ID: %v", deployment.DeploymentGroupID)
			return errors.New(errMsg)
		}

		deployment.DeploymentGroup = deploymentGroup
		deployment.DeploymentGroupName = deploymentGroup.Name
	} else if deployment.DeploymentGroupName != "" { // Try fetching by name
		deploymentGroup := FirstDeploymentGroup(
			"name = ? AND domain_id = ?",
			deployment.DeploymentGroupName,
			deployment.DomainID,
		)

		if deploymentGroup == nil {
			errMsg := fmt.Sprintf("No deployment group found with name: %v", deployment.DeploymentGroupName)
			return errors.New(errMsg)
		}

		deployment.DeploymentGroup = deploymentGroup
		deployment.DeploymentGroupID = &deploymentGroup.ID
	}

	return nil
}

//================================================================================
// Model Validations
//================================================================================

func (deployment *Deployment) validateRequiredParams() error {
	var missingParams []string

	if deployment.NodeCount == 0 {
		missingParams = append(missingParams, "node_count")
	}
	if deployment.ClusterName == "" {
		missingParams = append(missingParams, "cluster_name")
	}
	if deployment.State == "" {
		missingParams = append(missingParams, "state")
	}
	if deployment.Service == "" {
		missingParams = append(missingParams, "service")
	}
	if deployment.StorageProvider == "" {
		missingParams = append(missingParams, "storage_provider")
	}
	if _, cpuSet := deployment.Resources["cpu"]; !cpuSet {
		missingParams = append(missingParams, "resources.cpu")
	}
	if _, memSet := deployment.Resources["memory"]; !memSet {
		missingParams = append(missingParams, "resources.memory")
	}
	if _, diskSet := deployment.Resources["disk"]; !diskSet {
		missingParams = append(missingParams, "resources.disk")
	}
	if deployment.DatabaseTypeID == EmptyUUID && deployment.DatabaseTypeName == "" {
		missingParams = append(missingParams, "database_type_id or database_type")
	}
	if deployment.VersionID == EmptyUUID && deployment.VersionName == "" {
		missingParams = append(missingParams, "version_id or version")
	}
	if deployment.EnvironmentID == EmptyUUID && deployment.EnvironmentName == "" {
		missingParams = append(missingParams, "environment_id or environment")
	}
	if deployment.DomainID == EmptyUUID {
		missingParams = append(missingParams, "domain_id")
	}
	if deployment.ProjectID == EmptyUUID {
		missingParams = append(missingParams, "project_id")
	}
	if deployment.Origin != "legacy" {
		if deployment.TargetClusterID == nil && deployment.TargetClusterName == "" {
			missingParams = append(missingParams, "target_cluster_id or target_cluster")
		}
		if deployment.ImageID == nil {
			missingParams = append(missingParams, "image_id")
		}
	}

	if len(missingParams) > 0 {
		errMsg := fmt.Sprintf("Missing required param(s): %v", strings.Join(missingParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}

func (deployment *Deployment) validateParams() error {
	// Load the deployment associations if they haven't been loaded already
	if deployment.DatabaseType == nil {
		err := deployment.loadAssociations()
		if err != nil {
			return err
		}
	}

	// Validate backup schedules are supported by db type
	if !deployment.SupportsFullBackups() && deployment.FullBackupSchedule != "" {
		errMsg := fmt.Sprintf("%v does not support full backups.", deployment.DatabaseType.Name)
		return errors.New(errMsg)
	}
	if !deployment.SupportsIncrementalBackups() && deployment.IncrementalBackupSchedule != "" {
		errMsg := fmt.Sprintf("%v does not support incremental backups.", deployment.DatabaseType.Name)
		return errors.New(errMsg)
	}

	validServiceTypes := []string{"ClusterIP", "LoadBalancer", "NodePort", "ExternalName"}
	if deployment.ServiceType != "" && !contains(validServiceTypes, deployment.ServiceType) {
		return fmt.Errorf("invalid serviceType: %s, supported ones - %v", deployment.ServiceType, validServiceTypes)
	}

	return nil
}

//================================================================================
// Model Tranformation Helpers
//================================================================================

func (deployment *Deployment) marshalJSON() (err error) {
	// Configuration JSON
	var configuration []byte
	configuration, err = json.Marshal(deployment.Configuration)
	if err != nil {
		return err
	}
	deployment.ConfigurationJSON = postgres.Jsonb{configuration}

	// Resources JSON
	var resources []byte
	resources, err = json.Marshal(deployment.Resources)
	if err != nil {
		return err
	}
	deployment.ResourcesJSON = postgres.Jsonb{resources}

	// ConnectionDetails JSON
	if deployment.ConnectionDetails == nil {
		deployment.ConnectionDetails = make(map[string]interface{})
	}
	var connectionDetails []byte
	connectionDetails, err = json.Marshal(deployment.ConnectionDetails)
	if err != nil {
		return err
	}
	deployment.ConnectionDetailsJSON = postgres.Jsonb{connectionDetails}

	// Storage options JSON
	var storageOptions []byte
	storageOptions, err = json.Marshal(deployment.StorageOptions)
	if err != nil {
		return err
	}
	deployment.StorageOptionsJSON = postgres.Jsonb{storageOptions}

	return nil
}

func (deployment *Deployment) unmarshalJSON() (err error) {
	// Configuration JSON
	err = json.Unmarshal(deployment.ConfigurationJSON.RawMessage, &deployment.Configuration)
	if err != nil {
		return err
	}

	// Resources JSON
	err = json.Unmarshal(deployment.ResourcesJSON.RawMessage, &deployment.Resources)
	if err != nil {
		return err
	}

	// ConnectionDetails JSON
	if len(deployment.ConnectionDetailsJSON.RawMessage) == 0 {
		deployment.ConnectionDetails = make(map[string]interface{})
	} else {
		err = json.Unmarshal(deployment.ConnectionDetailsJSON.RawMessage, &deployment.ConnectionDetails)
		if err != nil {
			return err
		}
	}

	// Configuration JSON
	err = json.Unmarshal(deployment.StorageOptionsJSON.RawMessage, &deployment.StorageOptions)
	if err != nil {
		return err
	}

	return nil
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
