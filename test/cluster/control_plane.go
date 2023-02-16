package cluster

import (
	"context"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	pdsSystemNamespace = "pds-system"
)

// ControlPlane wraps a PDS control plane.
type ControlPlane struct {
	*cluster
}

// NewControlPlane creates a ControlPlane instance using the specified kubeconfig path.
// Fails if a kubernetes go-client cannot be configured based on the kubeconfig.
func NewControlPlane(kubeconfig string) (*ControlPlane, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	metaClient, err := metadata.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	cluster, err := newCluster(config, clientset, metaClient)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &ControlPlane{
		cluster: cluster,
	}, nil
}

// LogComponents extracts the logs of all relevant PDS components, beginning at the specified time.
func (cp *ControlPlane) LogComponents(t *testing.T, ctx context.Context, since time.Time) {
	t.Helper()
	components := []componentSelector{
		{pdsSystemNamespace, "component=api-server"},
		{pdsSystemNamespace, "component=api-worker"},
		{pdsSystemNamespace, "component=faktory"},
	}
	t.Log("Control plane:")
	cp.getLogsForComponents(t, ctx, components, since)
}
