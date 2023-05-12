package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationZookeeper is the database type
	ApplicationZookeeper string = "zookeeper"
	// ApplicationShortZookeeper the abbreviated database type
	ApplicationShortZookeeper string = "zk"
)

// ZookeeperSpec defines the desired state of Zookeeper
type ZookeeperSpec struct {
	DataServiceSpec `json:",inline"`
}

// ZookeeperStatus defines the observed state of Zookeeper
type ZookeeperStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails ZookeeperClusterDetails `json:"clusterDetails,omitempty"`
}

// ZookeeperClusterDetails provide the cluster details
type ZookeeperClusterDetails struct {
	ClusterName string `json:"clusterName,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	Port        int32  `json:"Port,omitempty"`
	Version     string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true

// Zookeeper is the Schema for the zookeepers API
// +kubebuilder:resource:shortName=zk;zks
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Zookeeper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZookeeperSpec   `json:"spec,omitempty"`
	Status ZookeeperStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ZookeeperList contains a list of Zookeeper
type ZookeeperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zookeeper `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Zookeeper{}, &ZookeeperList{})
}
