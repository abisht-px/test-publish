package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationPostgresql is the database type
	ApplicationPostgresql string = "postgresql"
	// ApplicationShortPostgresql the abbreviated database type
	ApplicationShortPostgresql string = "pg"
)

// PostgresqlSpec defines the desired state of Postgresql
type PostgresqlSpec struct {
	DataServiceSpec `json:",inline"`
}

// PostgresqlStatus defines the observed state of Postgresql
type PostgresqlStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails PostgresqlClusterDetails `json:"clusterDetails,omitempty"`
}

// PostgresqlClusterDetails provide the connection details
type PostgresqlClusterDetails struct {
	Host             string `json:"host,omitempty"`
	Port             int32  `json:"port,omitempty"`
	Version          string `json:"version,omitempty"`
	ConnectionString string `json:"connectionString,omitempty"`
}

//+kubebuilder:object:root=true

// Postgresql is the Schema for the postgresqls API
// +kubebuilder:resource:shortName=pg;pgs
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Postgresql struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresqlSpec   `json:"spec,omitempty"`
	Status PostgresqlStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PostgresqlList contains a list of Postgresql
type PostgresqlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Postgresql `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Postgresql{}, &PostgresqlList{})
}
