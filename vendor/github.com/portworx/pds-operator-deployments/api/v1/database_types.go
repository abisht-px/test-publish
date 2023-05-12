package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/external-dns/endpoint"
)

// SharedStorage is a structure encapsulating the storage that will be shared for the Database
type SharedStorage struct {
	StorageClass          storagev1.StorageClass       `json:"storageClass"`
	PersistentVolumeClaim corev1.PersistentVolumeClaim `json:"persistentVolumeClaim"`
}

// DatabaseSpec defines the desired state of Database
type DatabaseSpec struct {
	Type             string `json:"type"`
	Service          string `json:"service"`
	Application      string `json:"application"`
	ApplicationShort string `json:"applicationShort"`
	Environment      string `json:"environment"`
	// +optional
	RoleRules []rbacv1.PolicyRule `json:"roleRules,omitempty"`
	// +optional
	ClusterRoleRules []rbacv1.PolicyRule `json:"clusterRoleRules,omitempty"`
	// +optional
	Capabilities  map[string]string `json:"capabilities,omitempty"`
	ConfigMapData map[string]string `json:"configMapData"`
	// +optional
	Services []Service `json:"services,omitempty"`
	// +optional
	DNSEndpoints     []endpoint.DNSEndpoint           `json:"dnsEndpoints,omitempty"`
	DisruptionBudget policyv1.PodDisruptionBudgetSpec `json:"disruptionBudget"`
	//VolumePlacementStrategy
	// +optional
	SharedStorage SharedStorage          `json:"sharedStorage,omitempty"`
	StorageClass  storagev1.StorageClass `json:"storageClass"`
	StatefulSet   appsv1.StatefulSetSpec `json:"statefulSet"`
	// +optional
	Initialize InitializeType `json:"initialize,omitempty"`
	// +optional
	Mode DatabaseMode `json:"mode,omitempty"`
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// +optional
	TLSEnabled bool `json:"tlsEnabled,omitempty"`
	// +optional
	TLSIssuer string `json:"tlsIssuer,omitempty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	DataServiceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// Database is the Schema for the databases API
// +kubebuilder:resource:shortName=db;dbs
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.statefulSet.replicas,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.application`
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.statefulSet.replicas`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec,omitempty"`
	Status DatabaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}

func init() {
	// TODO: implement our own Endpoint custom resource
	SchemeBuilder.Register(&Database{}, &DatabaseList{})
}
