package cluster

import (
	"context"
	"testing"
	"time"
)

// TargetCluster wraps a PDS target cluster.
type TargetCluster struct {
	*cluster
}

// NewTargetCluster creates a TargetCluster instance with the specified kubeconfig.
// Fails if a kubernetes go-client cannot be configured based on the kubeconfig.
func NewTargetCluster(kubeconfig string) (*TargetCluster, error) {
	cluster, err := newCluster(kubeconfig)
	if err != nil {
		return nil, err
	}
	return &TargetCluster{cluster}, nil
}

// LogComponents extracts the logs of all relevant PDS components, beginning at the specified time.
func (tc *TargetCluster) LogComponents(t *testing.T, ctx context.Context, since time.Time) {
	t.Helper()
	components := []componentSelector{
		{pdsSystemNamespace, "app=pds-agent"},
		// TODO (fmilichovsky): Fix log extraction
		// (the operator pods consist of two containers, so this isn't enough to qualify the one we need).
		{pdsSystemNamespace, "control-plane=controller-manager"}, // Deployment + Backup operators.
	}
	t.Log("Target cluster:")
	tc.getLogsForComponents(t, ctx, components, since)
}
