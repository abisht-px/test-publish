package v1

import (
	deploymentsv1 "github.com/portworx/pds-operator-deployments/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupJobSpec defines the desired state of BackupJob
type BackupJobSpec struct {
	// +kubebuilder:validation:Required
	BackupName string `json:"backupName"`
	// +kubebuilder:validation:Required
	BackupSpec BackupSpec `json:"backupSpec"`
	// +kubebuilder:validation:Required
	DataServiceKind string `json:"dataServiceKind"`
	// +kubebuilder:validation:Required
	DataServiceName string `json:"dataServiceName"`
	// +kubebuilder:validation:Required
	DataServiceSpec deploymentsv1.DataServiceSpec `json:"dataServiceSpec"`
}

// BackupJobStatus defines the observed state of BackupJob
type BackupJobStatus struct {
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
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

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=bj
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.backup.type`
// +kubebuilder:printcolumn:name="Start Time",type=string,JSONPath=`.status.startTime`
// +kubebuilder:printcolumn:name="Completion Time",type=string,JSONPath=`.status.completionTime`
// +kubebuilder:printcolumn:name="Completion Status",type=string,JSONPath=`.status.completionStatus`
// +kubebuilder:printcolumn:name="CloudSnap ID",type=string,JSONPath=`.status.cloudSnapID`

// BackupJob is the Schema for the backupjobs API
type BackupJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupJobSpec   `json:"spec,omitempty"`
	Status BackupJobStatus `json:"status,omitempty"`
}

func (s *BackupJob) GetDeploymentID() string {
	return s.Spec.BackupSpec.DeploymentID
}

//+kubebuilder:object:root=true

// BackupJobList contains a list of BackupJob
type BackupJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupJob{}, &BackupJobList{})
}
