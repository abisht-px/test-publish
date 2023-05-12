package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationDatastaxEnterprise is the database type
	ApplicationDatastaxEnterprise string = "datastaxEnterprise"
	// ApplicationShortDatastaxEnterprise the abbreviated database type
	ApplicationShortDatastaxEnterprise string = "dse"
)

// DatastaxEnterpriseSpec defines the desired state of DatastaxEnterprise
type DatastaxEnterpriseSpec struct {
	DataServiceSpec `json:",inline"`
}

// DatastaxEnterpriseStatus defines the observed state of DatastaxEnterprise
type DatastaxEnterpriseStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails DatastaxEnterpriseClusterDetails `json:"clusterDetails,omitempty"`
}

// DatastaxEnterpriseClusterDetails provide the cluster details
type DatastaxEnterpriseClusterDetails struct {
	ClusterName string `json:"clusterName,omitempty"`
	Nodes       string `json:"nodes,omitempty"`
	CqlPort     int32  `json:"cqlPort,omitempty"`
	ThriftPort  int32  `json:"thriftPort,omitempty"`
	Version     string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true

// DatastaxEnterprise is the Schema for the datastaxenterprises API
// +kubebuilder:resource:shortName=dse;dses
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type DatastaxEnterprise struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatastaxEnterpriseSpec   `json:"spec,omitempty"`
	Status DatastaxEnterpriseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatastaxEnterpriseList contains a list of DatastaxEnterprise
type DatastaxEnterpriseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatastaxEnterprise `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatastaxEnterprise{}, &DatastaxEnterpriseList{})
}
