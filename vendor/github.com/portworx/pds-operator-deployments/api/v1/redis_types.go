package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationRedis is the database type
	ApplicationRedis string = "redis"
	// ApplicationShortRedis the abbreviated database type
	ApplicationShortRedis string = "red"
)

// RedisSpec defines the desired state of Redis
type RedisSpec struct {
	DataServiceSpec `json:",inline"`
}

// RedisStatus defines the observed state of Redis
type RedisStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails RedisClusterDetails `json:"clusterDetails,omitempty"`
}

// RedisClusterDetails provide the cluster details
type RedisClusterDetails struct {
	ClusterName string `json:"clusterName,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
	Port        int32  `json:"Port,omitempty"`
	Version     string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true

// Redis is the Schema for the redis API
// +kubebuilder:resource:shortName=red;reds
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedisList contains a list of Redis
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Redis{}, &RedisList{})
}
