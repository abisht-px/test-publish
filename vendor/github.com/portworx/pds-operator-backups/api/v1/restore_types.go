package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RestoreCompletionStatus string
type RestoreErrorCode string

// Enum of valid Restore Completion Statuses
const (
	// RestoreStatusInitial is the initial state when snapshot restore is initiated.
	RestoreStatusInitial RestoreCompletionStatus = ""
	// RestoreStatusPending means the restore has not yet started.
	RestoreStatusPending RestoreCompletionStatus = "Pending"
	// RestoringCloudSnap means we're waiting for the cloud snap restore to complete.
	RestoringCloudSnap RestoreCompletionStatus = "RestoringCloudSnap"
	// RestoringDataServiceCR means the restore PV and PVC resources are ready
	// and we're waiting to get the data service manifest from the backup.
	RestoringDataServiceCR RestoreCompletionStatus = "RestoringDataServiceCR"
	// RestoringDeployment means the new data service has been created and we're waiting for the restore process to complete.
	RestoringDeployment RestoreCompletionStatus = "RestoringDeployment"
	// DeploymentEnteringNormalMode means the restore process in data service has succeeded and we're waiting until it becomes healthy in normal mode.
	DeploymentEnteringNormalMode RestoreCompletionStatus = "DeploymentEnteringNormalMode"
	// RestoreStatusSuccessful means the restore succeeded.
	RestoreStatusSuccessful RestoreCompletionStatus = "Successful"
	// RestoreStatusFailed means the restore failed.
	RestoreStatusFailed RestoreCompletionStatus = "Failed"
)

const (
	// PXCloudCredentialsNotFound signals the cloud credentials specified in Restore spec were not found by Portworx.
	PXCloudCredentialsNotFound RestoreErrorCode = "PXCloudCredentialsNotFound"
	// PXTriggerRestoreFailed signals there was a problem triggering the cloudsnap restore in PX.
	PXTriggerRestoreFailed RestoreErrorCode = "PXCloudSnapRestoreTriggerFailed"
	// PXRestoreFailed signals there was a problem with finishing the cloudsnap restore in PX.
	PXRestoreFailed RestoreErrorCode = "PXCloudSnapRestoreFailed"
	// PXGetCloudSnapStatusFailed signals there was a problem with getting the cloudsnap restore status from PX.
	PXGetCloudSnapStatusFailed RestoreErrorCode = "PXGetCloudSnapStatusFailed"
	// PXGetVolumeFailed signals there was a problem with getting the volume from PX.
	PXGetVolumeFailed RestoreErrorCode = "PXGetVolumeFailed"
	// ReadDataServiceManifestFailed signals there was a problem with reading the data service manifest from the busybox pod.
	ReadDataServiceManifestFailed RestoreErrorCode = "ReadDataServiceManifestFailed"
	// UnmarshalDataServiceJSON signals there was a problem with unmarshalling the JSON with data service manifest.
	UnmarshalDataServiceJSON RestoreErrorCode = "UnmarshalDataServiceJSON"
)

// RestoreSpec defines the desired state of Restore
type RestoreSpec struct {
	// +kubebuilder:validation:Required
	DeploymentName string `json:"deploymentName"`
	// +kubebuilder:validation:Required
	CloudCredentialName string `json:"cloudCredentialName"`
	// +kubebuilder:validation:Required
	PXCloudSnapID string `json:"pxCloudSnapID"`
	// +kubebuilder:validation:Optional
	ImageVersion string `json:"imageVersion,omitempty"`
}

// RestoreStatus defines the observed state of Restore
type RestoreStatus struct {
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
	// +optional
	CompletionStatus RestoreCompletionStatus `json:"completionStatus,omitempty"`
	// Machine-readable code specifying why the restore failed. Must be filled when CompletionStatus is 'Failed'.
	// +optional
	ErrorCode RestoreErrorCode `json:"errorCode,omitempty"`
	// TODO DS-3886 Fill the status later.
	//// More detailed description related to the ErrorCode.
	//// This field is usually populated by messages from specific K8s resource where the restore process failed.
	//// +optional
	//ErrorMessage string `json:"errorMessage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Deployment Name",type=string,JSONPath=`.spec.deploymentName`
// +kubebuilder:printcolumn:name="CloudSnap ID",type=string,JSONPath=`.spec.pxCloudSnapID`
// +kubebuilder:printcolumn:name="Start Time",type=string,JSONPath=`.status.startTime`
// +kubebuilder:printcolumn:name="Completion Time",type=string,JSONPath=`.status.completionTime`
// +kubebuilder:printcolumn:name="Completion Status",type=string,JSONPath=`.status.completionStatus`

// Restore is the Schema for the restores API
type Restore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RestoreSpec   `json:"spec,omitempty"`
	Status RestoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RestoreList contains a list of Restore
type RestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Restore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Restore{}, &RestoreList{})
}
