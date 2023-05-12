package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationElasticsearch is the database type
	ApplicationElasticsearch string = "elasticsearch"
	// ApplicationShortElasticsearch the abbreviated database type
	ApplicationShortElasticsearch string = "es"
)

// ElasticsearchSpec defines the desired state of Elasticsearch
type ElasticsearchSpec struct {
	DataServiceSpec `json:",inline"`
}

// ElasticsearchStatus defines the observed state of Elasticsearch
type ElasticsearchStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails ElasticsearchClusterDetails `json:"clusterDetails,omitempty"`
}

// ElasticsearchClusterDetails provide the cluster details
type ElasticsearchClusterDetails struct {
	ClusterName string `json:"clusterName,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	Port        int32  `json:"Port,omitempty"`
	Version     string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true

// Elasticsearch is the Schema for the elasticsearches API
// +kubebuilder:resource:shortName=es;ess
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Elasticsearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchSpec   `json:"spec,omitempty"`
	Status ElasticsearchStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ElasticsearchList contains a list of Elasticsearch
type ElasticsearchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Elasticsearch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Elasticsearch{}, &ElasticsearchList{})
}
