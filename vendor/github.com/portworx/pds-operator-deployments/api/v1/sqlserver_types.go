package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationSqlserver is the database type
	ApplicationSqlserver string = "sqlserver"
	// ApplicationShortSqlserver the abbreviated database type
	ApplicationShortSqlserver string = "sql"
)

// SqlserverSpec defines the desired state of Sqlserver
type SqlserverSpec struct {
	DataServiceSpec `json:",inline"`
}

// SqlserverStatus defines the observed state of Sqlserver
type SqlserverStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails SqlserverClusterDetails `json:"clusterDetails,omitempty"`
}

// SqlserverClusterDetails provide the cluster details
type SqlserverClusterDetails struct {
	Hostname string `json:"hostname,omitempty"`
	Port     int32  `json:"Port,omitempty"`
	Version  string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true

// Sqlserver is the Schema for the sqlservers API
// +kubebuilder:resource:shortName=sql;sqls
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Sqlserver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SqlserverSpec   `json:"spec,omitempty"`
	Status SqlserverStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SqlserverList contains a list of Sqlserver
type SqlserverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sqlserver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sqlserver{}, &SqlserverList{})
}
