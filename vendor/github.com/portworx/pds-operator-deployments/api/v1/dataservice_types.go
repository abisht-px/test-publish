package v1

import (
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// PDSDomain TBD
	PDSDomain string = "pds.io"
	// PDSNamespace TBD
	PDSNamespace string = "deployments." + PDSDomain
	// PDSRegistry TBD
	PDSRegistry string = "registry-ds." + PDSDomain + "/pds/develop/"
	// PDSType TBD
	PDSType string = "dataservice"
	// PDSLabelService TBD
	PDSLabelService string = PDSNamespace + "/service"
	// PDSLabelEnvironment TBD
	PDSLabelEnvironment string = PDSNamespace + "/environment"
	// PDSLabelBuild TBD
	PDSLabelBuild string = PDSNamespace + "/build"
	// SCCNonRoot is the nonroot Security Context Constraints of OpenShift.
	SCCNonRoot = "nonroot"
	// SCCNonRootV2 is the nonroot-v2 Security Context Constraints of OpenShift introduced in OCP 4.11.
	SCCNonRootV2 = "nonroot-v2"
)

var (
	// DataServiceUserID TBD
	DataServiceUserID int64 = 30021
	// Default is fsGroupChangePolicy=OnRootMismatch
	DefaultFSGroupChangePolicy = corev1.FSGroupChangeOnRootMismatch
)

// DataServiceHealth is an enum of valid health statuses
type DataServiceHealth string

const (
	// DataServiceHealthHealthy is a healthy database status
	DataServiceHealthHealthy DataServiceHealth = "Healthy"
	// DataServiceHealthUnhealthy is an unhealthy database status
	DataServiceHealthUnhealthy DataServiceHealth = "Down"
	// DataServiceHealthDegraded is a degraded database status
	DataServiceHealthDegraded DataServiceHealth = "Degraded"
)

// InitializeType is an enum of valid Initialization types
// +kubebuilder:validation:Enum=Once;Always;Never;Manual
type InitializeType string

const (
	// InitializeTypeOnce initialize once
	InitializeTypeOnce InitializeType = "Once"
	// InitializeTypeAlways initialize always
	InitializeTypeAlways InitializeType = "Always"
	// InitializeTypeNever initialize never
	InitializeTypeNever InitializeType = "Never"
	// InitializeTypeManual initialize manual
	InitializeTypeManual InitializeType = "Manual"
)

// InitializedTypeStatus is an enum of valid Initialization types
type InitializedTypeStatus string

const (
	// InitializedTypeStatusYes initialized
	InitializedTypeStatusYes InitializedTypeStatus = "Yes"
	// InitializedTypeStatusNo not initialized
	InitializedTypeStatusNo InitializedTypeStatus = "No"
	// InitializedTypeStatusUnknown unknown
	InitializedTypeStatusUnknown InitializedTypeStatus = "Unknown"
	// InitializedTypeStatusPending unknown
	InitializedTypeStatusPending InitializedTypeStatus = "Pending"
	// InitializedTypeStatusManual unknown
	InitializedTypeStatusManual InitializedTypeStatus = "Manual"
)

// PDSModeEnvVarName is the name of the environment variable that defines the mode of the PDS.
// It should be present in all containers in our data service pods.
// The value of the variable should be one of the values of DatabaseMode enum.
const PDSModeEnvVarName = "PDS_MODE"

// DatabaseMode is an enum of valid Mode types
type DatabaseMode string

const (
	// PDSModeNormal normal mode
	PDSModeNormal DatabaseMode = "Normal"
	// PDSModeRestore restore mode
	PDSModeRestore DatabaseMode = "Restore"
	// PDSModePostRestore post restore mode
	PDSModePostRestore DatabaseMode = "PostRestore"
	// PDSModeMaintenance maintenance mode
	PDSModeMaintenance DatabaseMode = "Maintenance"
)

// Crons TBD
type Crons struct {
	// +optional
	Daily string `json:"daily,omitempty"`
	// +optional
	Weekly string `json:"weekly,omitempty"`
	// +optional
	Monthly string `json:"monthly,omitempty"`
}

// Service TBD
type Service struct {
	Name string `json:"name"`
	// +optional
	Publish string `json:"publish,omitempty"`
	// +optional
	Hostname string `json:"hostname,omitempty"`
	// +optional
	DNSZone string `json:"dnsZone,omitempty"`
	// +optional
	NoSelectors    bool `json:"noSelectors,omitempty"`
	corev1.Service `json:",inline"`
}

// StorageOptions TBD
type StorageOptions struct {
	// +optional
	Replicas string `json:"replicas,omitempty"`
	// +optional
	Group string `json:"group,omitempty"`
	// +optional
	ForceSpread string `json:"forceSpread,omitempty"`
	// +optional
	ReclaimPolicy corev1.PersistentVolumeReclaimPolicy `json:"reclaimPolicy,omitempty"`
	// +optional
	Filesystem string `json:"filesystem,omitempty"`
	// +optional
	Secure string `json:"secure,omitempty"`
}

func (o StorageOptions) GetFilesystem() string {
	if o.Filesystem == "" {
		return "xfs"
	}
	return o.Filesystem
}

// DataServiceSpec TBD
type DataServiceSpec struct {
	// +optional
	Image string `json:"image,omitempty"`
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	Version          string                        `json:"version"`
	// +optional
	VersionName string `json:"versionName,omitempty"`
	// +optional
	ImageBuild string `json:"imageBuild,omitempty"`
	// +optional
	Capabilities map[string]string `json:"capabilities,omitempty"`
	Nodes        *int32            `json:"nodes"`
	// +optional
	DNSZone string `json:"dnsZone,omitempty"`
	// +optional
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
	// +optional
	LoadBalancerSourceRanges []string               `json:"loadBalancerSourceRanges,omitempty"`
	StorageClass             storagev1.StorageClass `json:"storageClass"`
	// +optional
	StorageOptions StorageOptions              `json:"storageOptions,omitempty"`
	Resources      corev1.ResourceRequirements `json:"resources"`
	// +optional
	Configuration map[string]string `json:"configuration,omitempty"`
	// +optional
	InitializeType InitializeType `json:"initialize,omitempty"`
	// +optional
	Crons Crons `json:"crons,omitempty"`
	// +optional
	RestoreVolumeClaim *corev1.PersistentVolumeClaimVolumeSource `json:"restoreVolumeClaim,omitempty"`
	// +optional
	Mode DatabaseMode `json:"mode,omitempty"`
	// +optional
	TLSEnabled bool `json:"tlsEnabled,omitempty"`
	// +optional
	TLSIssuer string `json:"tlsIssuer,omitempty"`
}

// Version parameter is deprecated and replaced by a pair
// VersionName-ImageBuild. Here we support it for bw compatibility but
// let's not forget to remove it asap
// TODO: DS-4386
func (s *DataServiceSpec) GetVersion() string {
	return s.Version
}

// GetMode returns the mode of the database
func (s *DataServiceSpec) GetMode() DatabaseMode {
	if s.Mode == "" {
		return PDSModeNormal
	}
	return s.Mode
}

// PodInfo contains information about a pod.
// +structType=atomic
type PodInfo struct {
	// The IP of this pod.
	// +optional
	IP string `json:"ip,omitempty"`
	// Name is the Hostname of this pod.
	Name string `json:"name"`
	// Node hosting this pod.
	// +optional
	WorkerNode string `json:"workerNode,omitempty"`
}

// ConnectionDetails of data service.
type ConnectionDetails struct {
	// Nodes of the data service.
	// +optional
	Nodes []string `json:"nodes,omitempty"`
	// Ports provided by the data service (name and number).
	// +optional
	Ports map[string]int32 `json:"ports,omitempty"`
}

// ResourceConditions holds conditions of a K8s resource.
type ResourceConditions struct {
	// A managed PDS resource, e.g., a DB pod.
	// +optional
	Resource corev1.TypedLocalObjectReference `json:"resource,omitempty"`

	// Conditions of the resource.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// ResourceEvent is a struct that holds only reduced information from the regular corev1.Event.
type ResourceEvent struct {
	// Name of the original underlying Event.
	// +optional
	Name string `json:"name,omitempty"`

	// This should be a short, machine understandable string that gives the reason
	// for the transition into the object's current status.
	// +optional
	Reason string `json:"reason,omitempty"`

	// A human-readable description of the status of this operation.
	// +optional
	Message string `json:"message,omitempty"`

	// The time at which the most recent occurrence of this event was recorded.
	// +optional
	Timestamp metav1.Time `json:"timestamp,omitempty"`

	// Type of this event (Normal, Warning), new types could be added in the future.
	// +optional
	Type string `json:"type,omitempty"`

	// What action was taken/failed regarding to the Regarding object.
	// +optional
	Action string `json:"action,omitempty"`
}

// ResourceEvents holds events of a K8s resource.
type ResourceEvents struct {
	// A managed PDS resource, e.g., a DB stateful set.
	// +optional
	Resource corev1.TypedLocalObjectReference `json:"resource,omitempty"`

	// Events of the resource.
	// +optional
	Events []ResourceEvent `json:"events,omitempty"`
}

// DataServiceStatus TBD
type DataServiceStatus struct {
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
	// +optional
	Pods []PodInfo `json:"pods,omitempty"`
	// +optional
	NotReadyPods []PodInfo `json:"notReadyPods,omitempty"`
	// +optional
	Health DataServiceHealth `json:"health,omitempty"`
	// +optional
	Initialized InitializedTypeStatus `json:"initialized,omitempty"`
	// +optional
	ConnectionDetails ConnectionDetails `json:"connectionDetails,omitempty"`
	// Lists PDS managed resources and their conditions for debugging purposes.
	// +optional
	Resources []ResourceConditions `json:"resources,omitempty"`
	// Lists PDS managed resources and their events for debugging purposes.
	// +optional
	ResourceEvents []ResourceEvents `json:"resourceEvents,omitempty"`
}
