package targetcluster

import (
	"context"
	"fmt"
	"time"

	openstoragev1 "github.com/libopenstorage/operator/pkg/apis/core/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	PortworxCSIDriverName    = "pxd.portworx.com"
	PortworxInTreeDriverName = "kubernetes.io/portworx-volume"
)

func (tc *TargetCluster) GetPortworxCSIDriver(ctx context.Context) (*storagev1.CSIDriver, error) {
	return tc.GetCSIDriver(ctx, PortworxCSIDriverName)
}

func (tc *TargetCluster) ListPortworxClusterPods(ctx context.Context, namespace string) (*corev1.PodList, error) {
	return tc.ListPods(ctx, namespace, map[string]string{
		"name":                               "portworx",
		"storage":                            "true",
		"operator.libopenstorage.org/driver": "portworx",
	})
}

func (tc *TargetCluster) GetLatestStorageNodeOnlineTime(ctx context.Context) (time.Time, error) {
	storageNodes, err := tc.ListStorageNodes(ctx, tc.Portworx.Namespace)
	if err != nil {
		return time.Time{}, err
	}

	var onlineTime time.Time
	for _, storageNode := range storageNodes.Items {
		for _, cond := range storageNode.Status.Conditions {
			if cond.Type == openstoragev1.NodeStateCondition &&
				cond.Status == openstoragev1.NodeOnlineStatus &&
				cond.LastTransitionTime.Time.After(onlineTime) {
				onlineTime = cond.LastTransitionTime.Time
			}
		}
	}

	if onlineTime.IsZero() {
		return time.Time{}, fmt.Errorf("no storage node has transitioned to online")
	}

	return onlineTime, nil
}

func (tc *TargetCluster) WaitForHealthyPortworxCluster(ctx context.Context, t tests.T) {
	t.Helper()
	t.Log("Waiting for healthy Portworx cluster.")
	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		err := tc.IsPortworxClusterHealthy(ctx)
		assert.NoError(t, err)
	})
}

func (tc *TargetCluster) IsPortworxClusterHealthy(ctx context.Context) error {
	storageNodes, err := tc.ListStorageNodes(ctx, tc.Portworx.Namespace)
	if err != nil {
		return err
	}

	for _, storageNode := range storageNodes.Items {
		if storageNode.Status.Phase != string(openstoragev1.NodeOnlineStatus) {
			return fmt.Errorf("storage node %s is not online", storageNode.Name)
		}
	}

	pxClusterPods, err := tc.ListPortworxClusterPods(ctx, tc.Portworx.Namespace)
	if err != nil {
		return err
	}

	if len(storageNodes.Items) != len(pxClusterPods.Items) {
		return fmt.Errorf("storage nodes (%d) and portworx cluster pod (%d) count mismatch",
			len(storageNodes.Items), len(pxClusterPods.Items))
	}

	for _, pxClusterPod := range pxClusterPods.Items {
		if pxClusterPod.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("portworx cluster pod %s is not running", pxClusterPod.Name)
		}

		var ready bool
		for _, condition := range pxClusterPod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				ready = true
			}
		}

		if !ready {
			return fmt.Errorf("portworx cluster pod %s is not ready", pxClusterPod.Name)
		}
	}

	return nil
}

func (tc *TargetCluster) WaitForPortworxClusterRestart(ctx context.Context, t tests.T, referenceTime time.Time) {
	t.Helper()
	t.Logf("Waiting for Portworx cluster restart after reference time %v.", referenceTime)
	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		err := tc.HasPortworxClusterRestarted(ctx, referenceTime)
		assert.NoError(t, err)
	})
}

func (tc *TargetCluster) HasPortworxClusterRestarted(ctx context.Context, referenceTime time.Time) error {
	storageNodes, err := tc.ListStorageNodes(ctx, tc.Portworx.Namespace)
	if err != nil {
		return err
	}

	for _, storageNode := range storageNodes.Items {
		var onlineTime time.Time
		for _, cond := range storageNode.Status.Conditions {
			if cond.Type == openstoragev1.NodeStateCondition && cond.Status == openstoragev1.NodeOnlineStatus {
				onlineTime = cond.LastTransitionTime.Time
				break
			}
		}

		if onlineTime.IsZero() {
			return fmt.Errorf("storage node %s is not online (%s)", storageNode.Name, storageNode.Status.Phase)
		}
		if !onlineTime.After(referenceTime) {
			return fmt.Errorf("storage node %s has not transitioned to online (%v) after reference time (%v)",
				storageNode.Name, onlineTime, referenceTime)
		}
	}

	return tc.IsPortworxClusterHealthy(ctx)
}

func (tc *TargetCluster) MustSetStorageClusterCSIEnabled(ctx context.Context, t tests.T, enabled bool) {
	t.Helper()
	tc.WaitForHealthyPortworxCluster(ctx, t)

	sc, err := tc.FindStorageCluster(ctx)
	require.NoError(t, err)

	if sc.Spec.CSI.Enabled == enabled {
		return // Nothing to do.
	}

	referenceTransitionTime, err := tc.GetLatestStorageNodeOnlineTime(ctx)
	require.NoError(t, err)

	sc.Spec.CSI.Enabled = enabled
	_, err = tc.UpdateStorageCluster(ctx, sc)
	require.NoError(t, err)

	tc.WaitForPortworxClusterRestart(ctx, t, referenceTransitionTime)
}
