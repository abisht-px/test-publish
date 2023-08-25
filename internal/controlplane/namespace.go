package controlplane

import (
	"context"

	"github.com/stretchr/testify/require"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

// MustWaitForTestNamespace sets up a reference to the default namespace that will be used for test deployments.
func (c *ControlPlane) MustWaitForTestNamespace(ctx context.Context, t tests.T, name string) {
	namespace := c.MustWaitForNamespaceStatus(ctx, t, name, "available")
	require.NotNilf(t, namespace, "PDS test namespace %s is not available.", name)
	c.TestPDSNamespaceID = namespace.GetId()
}

func (c *ControlPlane) MustWaitForNamespaceStatus(ctx context.Context, t tests.T, name, expectedStatus string) *pds.ModelsNamespace {
	var (
		namespace *pds.ModelsNamespace
		err       error
	)
	wait.For(t, wait.ShortTimeout, wait.ShortRetryInterval, func(t tests.T) {
		namespace, err = c.PDS.GetNamespaceByName(ctx, c.testPDSDeploymentTargetID, name)
		require.NoErrorf(t, err, "Getting namespace %s.", name)
		require.NotNilf(t, namespace, "Could not find namespace %s.", name)
		require.Equalf(t, expectedStatus, namespace.GetStatus(), "Namespace %s not in status %s.", name, expectedStatus)
	})
	return namespace
}

func (c *ControlPlane) MustNeverGetNamespaceByName(ctx context.Context, t tests.T, name string) {
	require.Never(
		t,
		func() bool {
			namespace, err := c.PDS.GetNamespaceByName(ctx, c.testPDSDeploymentTargetID, name)
			return err != nil && namespace != nil
		},
		wait.QuickCheckTimeout, wait.ShortRetryInterval,
		"Namespace %s was not expected to be found in control plane.", name,
	)
}

func (c *ControlPlane) MustGetNamespaceForDeployment(ctx context.Context, t tests.T, deploymentID string) string {
	deployment, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespace, resp, err := c.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	return namespace.GetName()
}
