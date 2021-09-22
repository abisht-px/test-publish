package models

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"

	"github.anim.dreamworks.com/DreamCloud/stella-api/utils"
)

type Backup struct {
	Model
	FileSize        uint      `json:"file_size"` // FileSize is managed by the API
	JobHistoryLimit *uint     `json:"job_history_limit"`
	BackupTime      time.Time `json:"backup_time" binding:"-"` // BackupTime is  managed by the API
	DeploymentID    uuid.UUID `json:"deployment_id"`
	Build           string    `json:"-" gorm:"column:build"`                           // Build is auto generated
	BackupID        string    `json:"backup_id" gorm:"column:backup_id"`               // BackupID is auto generated
	DeploymentName  string    `json:"deployment" gorm:"column:deployment" binding:"-"` // DeploymentName will be auto filled based on provided DeploymentID
	State           string    `json:"state" mapstructure:"-"`                          // State is managed by the API
	BackupType      string    `json:"backup_type" mapstructure:"type" enums:"adhoc,scheduled"`
	BackupLevel     string    `json:"backup_level" mapstructure:"level" enums:"snapshot,incremental"`
	Schedule        string    `json:"schedule" mapstructure:"schedule"`
	Endpoint        string    `json:"endpoint" mapstructure:"-"` // Endpoint is populated by the Tekton pipeline on creation

	DomainID  uuid.UUID `json:"domain_id"`
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	UserLogin string    `json:"user_login"`

	Deployment *Deployment `json:"-" gorm:"-"`
}

// Complex Type Constants
func ValidBackupTypes() []string {
	return []string{"adhoc", "scheduled"}
}
func ValidBackupLevels() []string {
	return []string{"snapshot", "incremental"}
}

func NewBackup() Backup {
	return Backup{}
}

func AllBackups() []Backup {
	var backups []Backup
	db.Unscoped().Find(&backups)
	return backups
}

func BackupsCount(
	archived string,
	domainIDs interface{},
	projectIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
) int {
	var count int
	var backups []Backup

	sDB := dbWithScope(archived, domainIDs, projectIDs).Where(queryParams)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("deployment LIKE ?", searchParam)
	}

	sDB.Find(&backups).Count(&count)
	return count
}

func PaginatedBackups(
	archived string,
	domainIDs interface{},
	projectIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
	page int,
	perPage int,
	orderBy string,
) []Backup {
	var backups []Backup
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
		sDB = sDB.Where("deployment LIKE ?", searchParam)
	}

	sDB.Find(&backups)
	return backups
}

func FindBackup(id interface{}) (*Backup, error) {
	backup := NewBackup()
	resp := Find(&backup, id)

	if resp.Error != nil {
		return nil, resp.Error
	} else {
		err := backup.loadAssociations()
		return &backup, err
	}
}

func FirstBackup(query string, params ...interface{}) (*Backup, error) {
	// TODO PDS-325: Refactor to use db.First() instead
	backups, err := BackupsWhere(query, params...)
	if err != nil {
		return nil, err
	}
	if len(backups) == 0 {
		return nil, nil
	}
	return &backups[0], nil
}

func FirstBackupUnscoped(query string, params ...interface{}) (*Backup, error) {
	// TODO PDS-325: Refactor to use db.First() instead
	backups, err := AllBackupsWhere(query, params...)
	if err != nil {
		return nil, err
	}
	if len(backups) == 0 {
		return nil, nil
	}
	return &backups[0], nil
}

func BackupsWhere(query string, params ...interface{}) ([]Backup, error) {
	var backups []Backup
	db.Where(query, params...).Find(&backups)
	if err := loadBackupAssociations(&backups); err != nil {
		return backups, err
	}
	return backups, nil
}

func AllBackupsWhere(query string, params ...interface{}) ([]Backup, error) {
	var backups []Backup
	db.Unscoped().Where(query, params...).Find(&backups)
	if err := loadBackupAssociations(&backups); err != nil {
		return backups, err
	}
	return backups, nil
}

func CreateBackup(backup Backup) (*Backup, error) {
	if !db.NewRecord(backup) {
		err := backup.Save()
		return &backup, err
	}

	resp := db.Create(&backup)
	return &backup, resp.Error
}

func FindAndDeleteBackup(id interface{}) error {
	backup, err := FindBackup(id)
	if err != nil {
		return err
	}
	if backup.DeletedAt != nil {
		return nil // Record has already been deleted
	}
	return backup.Delete()
}

// Updates a single field
func (backup *Backup) Update(attrs ...interface{}) error {
	return db.Unscoped().Model(backup).Update(attrs...).Error
}

// Updates multiple fields
func (backup *Backup) UpdateFields(fields map[string]interface{}) error {
	return db.Unscoped().Model(backup).Updates(fields).Error
}

func (backup *Backup) Save() error {
	if db.NewRecord(*backup) {
		_, err := CreateBackup(*backup)
		return err
	}

	resp := db.Save(backup)
	return resp.Error
}

func (backup *Backup) Delete() error {
	if db.NewRecord(*backup) || backup.DeletedAt != nil {
		return nil // Record does not exist or has already been deleted.
	}

	resp := db.Delete(backup)
	return resp.Error
}

func (backup *Backup) Reload() error {
	b, err := FindBackup(backup.ID)
	if err != nil {
		return err
	}
	*backup = *b

	return nil
}

//================================================================================
// Helper Functions
//================================================================================

func (backup *Backup) ResourceURI() string {
	return fmt.Sprintf("/api/backups/%v", backup.ID)
}

func (backup *Backup) SupportsFullBackups() bool {
	val := backup.Deployment.DatabaseType.HasFullBackup
	return val != nil && *val
}

func (backup *Backup) SupportsIncrementalBackups() bool {
	val := backup.Deployment.DatabaseType.HasIncrementalBackup
	return val != nil && *val
}

//================================================================================
// Hooks
//================================================================================

func (backup *Backup) BeforeSave() (err error) {
	// Validate Required Params
	if err = backup.validateRequiredParams(); err != nil {
		return err
	}
	// Validate Associations Exist
	if err = backup.validateAssociations(); err != nil {
		return err
	}
	// Validate Params Config
	if err = backup.validateParams(); err != nil {
		return err
	}

	return nil
}

func (backup *Backup) BeforeCreate(scope *gorm.Scope) (err error) {
	// Call the base Model's BeforeCreate
	if err = backup.Model.BeforeCreate(scope); err != nil {
		return err
	}
	// Set Defaults
	return backup.setDefaults()
}

func (backup *Backup) AfterUpdate() error {
	return backup.updateAssociations()
}

func (backup *Backup) AfterDelete() error {
	return backup.removeAssociations()
}

//================================================================================
// Associations
//================================================================================

// --- Loading Associations ---

func loadBackupAssociations(backups *[]Backup) error {
	for idx := range *backups {
		if err := (*backups)[idx].loadAssociations(); err != nil {
			return err
		}
	}
	return nil
}

func (backup *Backup) loadAssociations() error {
	return backup.loadDeployment()
}

func (backup *Backup) loadDeployment() (err error) {
	backup.Deployment, err = FindDeployment(backup.DeploymentID)
	return err
}

// --- Validating Associations ---

func (backup *Backup) validateAssociations() (err error) {
	if backup.Deployment == nil {
		if err = backup.loadDeployment(); err != nil {
			return err
		}
	}
	// Set DeploymentName
	backup.DeploymentName = backup.Deployment.DeploymentID
	return nil
}

// --- Updating Associations ---

func (backup *Backup) updateAssociations() error {
	if backup.BackupType != "scheduled" {
		return nil // Only need to update deployments for scheduled backups
	}

	deployment, err := FindDeployment(backup.DeploymentID)
	if err != nil {
		return err
	}

	needsUpdate := false

	if backup.BackupLevel == "snapshot" {
		if deployment.FullBackupSchedule != backup.Schedule {
			deployment.FullBackupSchedule = backup.Schedule
			needsUpdate = true
		}
		if deployment.FullBackupLimit != backup.JobHistoryLimit {
			deployment.FullBackupLimit = backup.JobHistoryLimit
			needsUpdate = true
		}
	} else if backup.BackupLevel == "incremental" {
		if deployment.IncrementalBackupSchedule != backup.Schedule {
			deployment.IncrementalBackupSchedule = backup.Schedule
			needsUpdate = true
		}
		if deployment.IncrementalBackupLimit != backup.JobHistoryLimit {
			deployment.IncrementalBackupLimit = backup.JobHistoryLimit
			needsUpdate = true
		}
	}

	if needsUpdate {
		err = deployment.Save()
	}

	return err
}

func (backup *Backup) removeAssociations() error {
	if backup.BackupType != "scheduled" {
		return nil // Only need to update deployments for scheduled backups
	}

	deployment, err := FindDeployment(backup.DeploymentID)
	if err != nil {
		return err
	}

	if backup.BackupLevel == "snapshot" {
		deployment.FullBackupSchedule = ""
		deployment.FullBackupLimit = nil
	} else if backup.BackupLevel == "incremental" {
		deployment.IncrementalBackupSchedule = ""
		deployment.IncrementalBackupLimit = nil
	}

	return deployment.Save()
}

//================================================================================
// Model Validations
//================================================================================

func (backup *Backup) validateRequiredParams() error {
	var missingParams []string

	// Validate required parameters
	if backup.DeploymentID == EmptyUUID {
		missingParams = append(missingParams, "deployment_id")
	}
	if len(missingParams) > 0 {
		errMsg := fmt.Sprintf("Missing required param(s): %v", strings.Join(missingParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}

func (backup *Backup) validateParams() error {
	// Load the backup associations if they haven't been loaded already
	if backup.Deployment == nil {
		if err := backup.loadAssociations(); err != nil {
			return err
		}
	}

	// Validate allowed parameter values
	if !utils.ContainsString(ValidBackupTypes(), backup.BackupType) {
		errMsg := fmt.Sprintf("Param backup_type must be one of: %v", ValidBackupTypes())
		return errors.New(errMsg)
	}
	if !utils.ContainsString(ValidBackupLevels(), backup.BackupLevel) {
		errMsg := fmt.Sprintf("Param backup_level must be one of: %v", ValidBackupLevels())
		return errors.New(errMsg)
	}

	// Validate db type supports backup type
	if backup.BackupLevel == "snapshot" && !backup.SupportsFullBackups() {
		errMsg := fmt.Sprintf("%v does not support full backups.", backup.Deployment.DatabaseType.Name)
		return errors.New(errMsg)
	}
	if backup.BackupLevel == "incremental" && !backup.SupportsIncrementalBackups() {
		errMsg := fmt.Sprintf("%v does not support incremental backups.", backup.Deployment.DatabaseType.Name)
		return errors.New(errMsg)
	}

	return nil
}

//================================================================================
// Defaults
//================================================================================

func (backup *Backup) setDefaults() error {
	// Load the backup associations if they haven't been loaded already
	if backup.Deployment == nil {
		if err := backup.loadAssociations(); err != nil {
			return err
		}
	}

	// Set Project and Domain IDs based on the Deployment
	backup.DomainID = backup.Deployment.DomainID
	backup.ProjectID = backup.Deployment.ProjectID

	// Set BackupTime to current time if not provided
	if backup.BackupTime == EmptyTime {
		backup.SetBackupTime()
	}

	// Set Build and Backup ID
	build := ""
	for {
		seed := fmt.Sprintf("%v-%v-%v", backup.Deployment.ClusterName, backup.Deployment.EnvironmentName, time.Now())
		sha := fmt.Sprintf("%x", sha256.Sum256([]byte(seed)))
		build = utils.Substring(sha, 0, 8)

		backup, err := FirstBackup("build = ?", build)
		if err != nil {
			return err
		}
		if backup == nil {
			break // build is unique
		}
	}
	backup.Build = build
	backup.BackupID = fmt.Sprintf("backup-%v-%v", build, backup.Deployment.DeploymentID)

	// PDS-270 investigate whether we need to store the endpoint.
	backup.Endpoint = fmt.Sprintf("%s/apis/backups.pds.io/v1/namespaces/%s/backups/%s",
		backup.Deployment.TargetCluster.APIServer,
		backup.Deployment.EnvironmentName,
		backup.BackupID)
	return nil
}

func (backup *Backup) SetBackupTime() {
	backup.BackupTime = time.Now()
}
