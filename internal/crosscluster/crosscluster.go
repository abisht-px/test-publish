package crosscluster

import (
	"time"

	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
)

const (
	waiterRetryInterval = time.Second * 10
)

// CrossClusterHelper defines helper functions that involve both the control plane and target cluster.
// If a check can be performed exclusively in the context of the API/control plane/target cluster, it doesn't belong
// in CrossClusterHelper.
//
// It deliberately lacks access to a testing.T instance, forcing callers to supply their own (avoiding parallel subtest issues).
type CrossClusterHelper struct {
	controlPlane  *controlplane.ControlPlane
	targetCluster *targetcluster.TargetCluster

	startTime time.Time
}

func NewHelper(controlPlane *controlplane.ControlPlane, targetCluster *targetcluster.TargetCluster, startTime time.Time) *CrossClusterHelper {
	return &CrossClusterHelper{
		controlPlane:  controlPlane,
		targetCluster: targetCluster,
		startTime:     startTime,
	}
}
