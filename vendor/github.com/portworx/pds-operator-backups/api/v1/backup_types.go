package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BackupJobCompletionStatus string
type BackupJobErrorCode string
type ReclaimPolicy string

// Enum of valid Job Completion Statuses
const (
	// BackupJobActive means the job has not yet completed
	BackupJobActive BackupJobCompletionStatus = "Active"
	// BackupJobSucceeded means the job has completed successfully
	BackupJobSucceeded BackupJobCompletionStatus = "Succeeded"
	// BackupJobFailed means the job has completed unsuccessfully
	BackupJobFailed BackupJobCompletionStatus = "Failed"

	// Retain retains the VolumeSnapshots for a backup object
	RetainVS ReclaimPolicy = "retain"
	// Delete deletes the VolumeSnapshots for a backup object
	DeleteVS ReclaimPolicy = "delete"
)

// Enum of valid BackupJob Error Codes.
const (
	// JobFailedErrorCode means the underlying Job failed.
	JobFailedErrorCode BackupJobErrorCode = "JobFailed"
	// VolumeSnapshotFailedErrorCode means the associated VolumeSnapshot failed to get synced to the object store.
	VolumeSnapshotFailedErrorCode BackupJobErrorCode = "VolumeSnapshotFailed"
	// PXCloudCredentialsNotFoundErrorCode means the associated PX Credentials could not be found in the PX cluster.
	PXCloudCredentialsNotFoundErrorCode BackupJobErrorCode = "PXCloudCredentialsNotFound"
	// VolumeSnapshotCreationErrorCode means that there was a problem during volume creation.
	VolumeSnapshotCreationErrorCode BackupJobErrorCode = "VolumeSnapshotCreationError"
	// NoAssociatedDeploymentErrorCode means that the deployment backup was supposed to be performed on, does not exist.
	NoAssociatedDeploymentErrorCode BackupJobErrorCode = "NoAssociatedDeployment"
)

var (
	PDSUserID int64 = 30021
)

// InlineBackupJobStatus defines the observed state of a BackupJob
type InlineBackupJobStatus struct {
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	CompletionStatus BackupJobCompletionStatus `json:"completionStatus,omitempty"`
	// Machine-readable code specifying why the backup job failed. Must be filled when CompletionStatus is 'Failed'.
	// +optional
	ErrorCode BackupJobErrorCode `json:"errorCode,omitempty"`
	// More detailed description related to the ErrorCode.
	// This field is usually populated by messages from specific K8s resource where the backup job process failed.
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
	// ID of the Portworx CloudSnap necessary for the restore operation.
	// +optional
	CloudSnapID string `json:"cloudSnapID,omitempty"`
}

// BackupSpec defines the desired state of Backup
type BackupSpec struct {
	// +kubebuilder:validation:Required
	DeploymentID string `json:"deploymentID,omitempty"`
	// +kubebuilder:validation:Enum=adhoc;scheduled
	// //+kubebuilder:default=adhoc
	Type string `json:"type,omitempty"`
	// +kubebuilder:validation:Enum=snapshot;incremental
	// //+kubebuilder:default=snapshot
	Level string `json:"level,omitempty"`
	// TODO: create a validating admission webhook to validate
	//       that a Schedule is provided if Type=scheduled
	// +kubebuilder:validation:Optional
	Schedule string `json:"schedule,omitempty"`
	// //TODO: +kubebuilder:default=10
	// +kubebuilder:validation:Optional
	JobHistoryLimit *int32 `json:"jobHistoryLimit,omitempty"`
	// +kubebuilder:validation:Optional:default:=docker.io/openstorage/
	ImageRegistry string `json:"imageRegistry,omitempty"`
	// +kubebuilder:validation:Optional:default:=false
	Suspend bool `json:"suspend,omitempty"`
	// +kubebuilder:validation:Enum=retain;delete
	ReclaimPolicy ReclaimPolicy `json:"reclaimPolicy,omitempty"`
	// +kubebuilder:validation:Optional:default:=pds-backups
	CloudCredentialName string `json:"cloudCredentialName,omitempty"`
}

// BackupStatus defines the observed state of Backup
type BackupStatus struct {
	// +optional
	LastStartTime *metav1.Time `json:"lastStartTime,omitempty"`
	// +optional
	LastCompletionTime *metav1.Time `json:"lastCompletionTime,omitempty"`
	// +optional
	NextStartTime *metav1.Time `json:"nextStartTime,omitempty"`
	// +optional
	Active int32 `json:"active,omitempty"`
	// +optional
	Succeeded int32 `json:"succeeded,omitempty"`
	// +optional
	Failed int32 `json:"failed,omitempty"`
	// +optional
	BackupJobs []InlineBackupJobStatus `json:"backupJobs,omitempty"`
	// +optional
	Error string `json:"error,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="DeploymentID",type=string,JSONPath=`.spec.deploymentID`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Level",type=string,JSONPath=`.spec.level`
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`
// +kubebuilder:printcolumn:name="Last Start Time",type=string,JSONPath=`.status.lastStartTime`
// +kubebuilder:printcolumn:name="Last Completion Time",type=string,JSONPath=`.status.lastCompletionTime`
// +kubebuilder:printcolumn:name="Next Start Time",type=string,JSONPath=`.status.nextStartTime`
// +kubebuilder:printcolumn:name="Active",type=string,JSONPath=`.status.active`
// +kubebuilder:printcolumn:name="Succeeded",type=string,JSONPath=`.status.succeeded`
// +kubebuilder:printcolumn:name="Failed",type=string,JSONPath=`.status.failed`
// +kubebuilder:printcolumn:name="Error", type=string,JSONPath=`.status.error`

// Backup is the Schema for the backups API
type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupSpec   `json:"spec,omitempty"`
	Status BackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackupList contains a list of Backup
type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backup{}, &BackupList{})
}
