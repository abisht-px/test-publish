package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KafkaSpec defines the desired state of Kafka
type KafkaSpec struct {
	DataServiceSpec `json:",inline"`
}

// KafkaStatus defines the observed state of Kafka
type KafkaStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails KafkaClusterDetails `json:"clusterDetails,omitempty"`
}

// KafkaClusterDetails provide the cluster details
type KafkaClusterDetails struct {
	ClusterName string `json:"clusterName,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	Port        int32  `json:"Port,omitempty"`
	Version     string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true

// Kafka is the Schema for the kafkas API
// +kubebuilder:resource:shortName=kf;kfs
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Kafka struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaSpec   `json:"spec,omitempty"`
	Status KafkaStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KafkaList contains a list of Kafka
type KafkaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kafka `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kafka{}, &KafkaList{})
}

const (
	// ApplicationKafka is the database type
	ApplicationKafka string = "kafka"
	// ApplicationShortKafka the abbreviated database type
	ApplicationShortKafka string = "kf"
)
