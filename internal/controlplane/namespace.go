package controlplane

import (
	"context"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	waiterNamespaceExistsTimeout = time.Second * 30
)

// MustWaitForTestNamespace sets up a reference to the default namespace that will be used for test deployments.
func (c *ControlPlane) MustWaitForTestNamespace(ctx context.Context, t tests.T, name string) {
	namespace := c.MustWaitForNamespaceStatus(ctx, t, name, "available")
	require.NotNilf(t, namespace, "PDS test namespace %s is not available.", name)
	c.testPDSNamespaceID = namespace.GetId()
}

func (c *ControlPlane) MustWaitForNamespaceStatus(ctx context.Context, t tests.T, name, expectedStatus string) *pds.ModelsNamespace {
	var (
		namespace *pds.ModelsNamespace
		err       error
	)
	wait.For(t, waiterNamespaceExistsTimeout, waiterShortRetryInterval, func(t tests.T) {
		namespace, err = c.API.GetNamespaceByName(ctx, c.testPDSDeploymentTargetID, name)
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
			namespace, err := c.API.GetNamespaceByName(ctx, c.testPDSDeploymentTargetID, name)
			return err != nil && namespace != nil
		},
		waiterNamespaceExistsTimeout, waiterShortRetryInterval,
		"Namespace %s was not expected to be found in control plane.", name,
	)
}
