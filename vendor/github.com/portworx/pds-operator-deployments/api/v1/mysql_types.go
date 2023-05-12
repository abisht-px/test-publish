package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ApplicationMysql is the database type
	ApplicationMysql string = "mysql"
	// ApplicationShortMysql the abbreviated database type
	ApplicationShortMysql string = "my"
)

// MysqlSpec defines the desired state of Mysql
type MysqlSpec struct {
	DataServiceSpec `json:",inline"`
}

// MysqlStatus defines the observed state of Mysql
type MysqlStatus struct {
	*DataServiceStatus `json:",inline"`
	// +optional
	ClusterDetails MysqlClusterDetails `json:"clusterDetails,omitempty"`
}

// MysqlClusterDetails provide the cluster details
type MysqlClusterDetails struct {
	Hostname string `json:"hostname,omitempty"`
	Port     int32  `json:"Port,omitempty"`
	Version  string `json:"version,omitempty"`
}

// +kubebuilder:object:root=true

// Mysql is the Schema for the mysqls API
// +kubebuilder:resource:shortName=my;mys
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.nodes,statuspath=.status.replicas
// +kubebuilder:printcolumn:name="Environment",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/environment`
// +kubebuilder:printcolumn:name="Service",type=string,JSONPath=`.metadata.labels.deployments\.pds\.io\/service`
// +kubebuilder:printcolumn:name="Desired",type=integer,JSONPath=`.spec.nodes`
// +kubebuilder:printcolumn:name="Current",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Health",type=string,JSONPath=`.status.health`
// +kubebuilder:printcolumn:name="Initialized",type=string,JSONPath=`.status.initialized`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Mysql struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlSpec   `json:"spec,omitempty"`
	Status MysqlStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MysqlList contains a list of Mysql
type MysqlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mysql `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mysql{}, &MysqlList{})
}
