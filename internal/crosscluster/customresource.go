package crosscluster

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (c *CrossClusterHelper) MustDeleteDeploymentCustomResource(ctx context.Context, t tests.T, deploymentId string, database string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentId).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)
	namespace := namespaceModel.GetName()

	customResourceName := *deployment.ClusterResourceName

	err = c.targetCluster.DeletePDSDeployment(ctx, namespace, database, customResourceName)
	require.NoError(t, err)

	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		_, err := c.targetCluster.GetPDSDeployment(ctx, namespace, database, customResourceName)
		expectedError := fmt.Sprintf("%s.deployments.pds.io %q not found", database, customResourceName)
		require.EqualError(t, err, expectedError, "deployment CR is not deleted.")
	})
}
