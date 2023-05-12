package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationRabbitmq is the database type
	ApplicationRabbitmq string = "rabbitmq"
	// ApplicationShortRabbitmq the abbreviated database type
	ApplicationShortRabbitmq string = "rmq"
)

// RabbitmqSpec defines the desired state of Rabbitmq
type RabbitmqSpec struct {
	DataServiceSpec `json:",inline"`
}

// RabbitmqStatus defines the observed state of Rabbitmq
type RabbitmqStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails RabbitmqClusterDetails `json:"clusterDetails,omitempty"`
}

// RabbitmqClusterDetails provide the cluster details
type RabbitmqClusterDetails struct {
	Host             string `json:"host,omitempty"`
	Port             int32  `json:"port,omitempty"`
	Version          string `json:"version,omitempty"`
	ConnectionString string `json:"connectionString,omitempty"`
}

// +kubebuilder:object:root=true

// Rabbitmq is the Schema for the rabbitmqs API
// +kubebuilder:resource:shortName=rmq;rmqs
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Rabbitmq struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitmqSpec   `json:"spec,omitempty"`
	Status RabbitmqStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RabbitmqList contains a list of Rabbitmq
type RabbitmqList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rabbitmq `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rabbitmq{}, &RabbitmqList{})
}
