package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationCouchbase is the database type
	ApplicationCouchbase string = "couchbase"
	// ApplicationShortCouchbase the abbreviated database type
	ApplicationShortCouchbase string = "cb"
)

// CouchbaseSpec defines the desired state of Couchbase
type CouchbaseSpec struct {
	DataServiceSpec `json:",inline"`
}

// CouchbaseStatus defines the observed state of Couchbase
type CouchbaseStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails CouchbaseClusterDetails `json:"clusterDetails,omitempty"`
}

// CouchbaseClusterDetails provide the cluster details
type CouchbaseClusterDetails struct {
	ClusterName string `json:"clusterName,omitempty"`
	Host        string `json:"host,omitempty"`
	Port        int32  `json:"port,omitempty"`
	Bucket      string `json:"bucket,omitempty"`
	Version     string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true

// Couchbase is the Schema for the couchbases API
// +kubebuilder:resource:shortName=cb;cbs
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.status.replicas`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Couchbase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CouchbaseSpec   `json:"spec,omitempty"`
	Status CouchbaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CouchbaseList contains a list of Couchbase
type CouchbaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Couchbase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Couchbase{}, &CouchbaseList{})
}
