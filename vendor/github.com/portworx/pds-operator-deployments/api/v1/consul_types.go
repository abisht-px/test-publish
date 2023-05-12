package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationConsul is the database type
	ApplicationConsul string = "consul"
	// ApplicationShortConsul the abbreviated database type
	ApplicationShortConsul string = "con"
)

// ConsulSpec defines the desired state of Consul
type ConsulSpec struct {
	DataServiceSpec `json:",inline"`
}

// ConsulStatus defines the observed state of Consul
type ConsulStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails ConsulClusterDetails `json:"clusterDetails,omitempty"`
}

// ConsulClusterDetails provide the cluster details
type ConsulClusterDetails struct {
	ClusterName   string `json:"clusterName,omitempty"`
	Nodes         int32  `json:"nodes,omitempty"`
	Endpoints     string `json:"endpoints,omitempty"`
	DNSPort       int32  `json:"dnsPort,omitempty"`
	HTTPPort      int32  `json:"httpPort,omitempty"`
	ClientRPCPort int32  `json:"clientRpcPort,omitempty"`
	ServerRPCPort int32  `json:"serverRpcPort,omitempty"`
	UI            string `json:"ui,omitempty"`
	Version       string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true

// Consul is the Schema for the consuls API
// +kubebuilder:resource:shortName=con;cons
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Consul struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsulSpec   `json:"spec,omitempty"`
	Status ConsulStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConsulList contains a list of Consul
type ConsulList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Consul `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Consul{}, &ConsulList{})
}
