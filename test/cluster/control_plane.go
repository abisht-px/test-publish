package cluster

import (
	"context"
	"testing"
	"time"
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
	cluster, err := newCluster(kubeconfig)
	if err != nil {
		return nil, err
	}
	return &ControlPlane{cluster}, nil
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
