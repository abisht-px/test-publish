package crosscluster

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (c *CrossClusterHelper) MustWaitForDeploymentInitialized(ctx context.Context, t tests.T, deploymentID string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	clusterInitJobName := fmt.Sprintf("%s-cluster-init", deployment.GetClusterResourceName())
	nodeInitJobName := fmt.Sprintf("%s-node-init", deployment.GetClusterResourceName())

	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		clusterInitJob, err := c.targetCluster.GetJob(ctx, namespace, clusterInitJobName)
		require.NoErrorf(t, err, "Getting clusterInitJob %s/%s for deployment %s.", namespace, clusterInitJobName, deploymentID)
		require.Truef(t, isJobSucceeded(clusterInitJob), "ClusterInitJob %s/%s for deployment %s not successful.", namespace, clusterInitJobName, deploymentID)

		nodeInitJob, err := c.targetCluster.GetJob(ctx, namespace, nodeInitJobName)
		require.NoErrorf(t, err, "Getting nodeInitJob %s/%s for deployment %s.", namespace, nodeInitJobName, deploymentID)
		require.Truef(t, isJobSucceeded(clusterInitJob), "NodeInitJob %s/%s for deployment %s not successful.", namespace, nodeInitJob, deploymentID)
	})
}

func isJobSucceeded(job *batchv1.Job) bool {
	return *job.Spec.Completions == job.Status.Succeeded
}
