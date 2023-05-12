package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationMongodb is the database type
	ApplicationMongodb string = "mongodb"
	// ApplicationShortMongodb the abbreviated database type
	ApplicationShortMongodb string = "mdb"
)

// MongodbSpec defines the desired state of Mongodb
type MongodbSpec struct {
	DataServiceSpec `json:",inline"`
}

// MongodbStatus defines the observed state of Mongodb
type MongodbStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails MongodbClusterDetails `json:"clusterDetails,omitempty"`
}

// MongodbClusterDetails provide the cluster details
type MongodbClusterDetails struct {
	Host             string `json:"host,omitempty"`
	Port             int32  `json:"port,omitempty"`
	Version          string `json:"version,omitempty"`
	ConnectionString string `json:"connectionString,omitempty"`
}

// +kubebuilder:object:root=true

// Mongodb is the Schema for the mongodbs API
// +kubebuilder:resource:shortName=mdb;mdbs
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Mongodb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongodbSpec   `json:"spec,omitempty"`
	Status MongodbStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MongodbList contains a list of Mongodb
type MongodbList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mongodb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mongodb{}, &MongodbList{})
}
