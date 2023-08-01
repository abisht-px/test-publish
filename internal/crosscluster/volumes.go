package crosscluster

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/portworx/pds-integration-test/internal/tests"
)

const (
	pdsDeploymentIDLabel = "pds/deployment-id"
)

// MustDeleteDeploymentVolumes deletes Persistent Volume Clames and it's volumes.
// Doesn't fail on error, can do clean up with helper scripts.
func (c *CrossClusterHelper) MustDeleteDeploymentVolumes(ctx context.Context, t tests.T, deploymentID string) {
	pvList, err := c.targetCluster.Clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", pdsDeploymentIDLabel, deploymentID),
	})
	if err != nil {
		t.Logf("failed to list %s database PersistentVolumes: %s", deploymentID, err)
		return
	}

	for _, pv := range pvList.Items {
		// Delete PersistentVolumeClaim
		err = c.targetCluster.Clientset.CoreV1().PersistentVolumeClaims(pv.Spec.ClaimRef.Namespace).Delete(ctx, pv.Spec.ClaimRef.Name, metav1.DeleteOptions{})
		if err != nil {
			t.Logf("delete %s/%s PersistentVolumeClaim: %s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name, err)
		}

		// Delete PersistentVolume.
		err = c.targetCluster.Clientset.CoreV1().PersistentVolumes().Delete(ctx, pv.GetName(), metav1.DeleteOptions{})
		if err != nil {
			t.Logf("delete %s PersistentVolume: %s", pv.GetName(), err)
		}
	}
}
