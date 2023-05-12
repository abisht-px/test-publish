package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationCassandra is the database type
	ApplicationCassandra string = "cassandra"
	// ApplicationShortCassandra the abbreviated database type
	ApplicationShortCassandra string = "cas"
)

// CassandraSpec defines the desired state of Cassandra
type CassandraSpec struct {
	DataServiceSpec `json:",inline"`
}

// CassandraStatus defines the observed state of Cassandra
type CassandraStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails CassandraClusterDetails `json:"clusterDetails,omitempty"`
}

// CassandraClusterDetails provide the cluster details
type CassandraClusterDetails struct {
	ClusterName string `json:"clusterName,omitempty"`
	Nodes       string `json:"nodes,omitempty"`
	CqlPort     int32  `json:"cqlPort,omitempty"`
	ThriftPort  int32  `json:"thriftPort,omitempty"`
	Version     string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true

// Cassandra is the Schema for the cassandras API
// +kubebuilder:resource:shortName=cas;cass
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraSpec   `json:"spec,omitempty"`
	Status CassandraStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CassandraList contains a list of Cassandra
type CassandraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cassandra `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cassandra{}, &CassandraList{})
}
