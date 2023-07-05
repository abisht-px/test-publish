package controlplane

import (
	"context"
	"net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (c *ControlPlane) MustCreateRestore(ctx context.Context, t tests.T, backupJobID, name, nsID, deploymentTargetID string) *pds.ModelsRestore {
	restore, resp, err := c.CreateRestore(ctx, backupJobID, name, nsID, deploymentTargetID)
	api.RequireNoError(t, resp, err)
	require.NotNil(t, restore)
	return restore
}

func (c *ControlPlane) CreateRestore(ctx context.Context, backupJobID, name, namespaceID, deploymentTargetID string) (*pds.ModelsRestore, *http.Response, error) {
	requestBody := pds.RequestsCreateRestoreRequest{
		Name:               &name,
		NamespaceId:        &namespaceID,
		DeploymentTargetId: &deploymentTargetID,
	}

	return c.PDS.RestoresApi.ApiBackupJobsIdRestorePost(ctx, backupJobID).Body(requestBody).Execute()
}

func (c *ControlPlane) MustWaitForRestoreSuccessful(ctx context.Context, t tests.T, restoreID string) {
	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		restore, resp, err := c.PDS.RestoresApi.ApiRestoresIdGet(ctx, restoreID).Execute()
		api.RequireNoError(t, resp, err)
		state := restore.GetStatus()
		require.Equal(t, "Successful", state, "Restore %q is in state %q.", restoreID, state)
	})
}

func (c *ControlPlane) MustWaitForRestoreFailed(ctx context.Context, t tests.T, restoreID string) {
	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		restore, resp, err := c.PDS.RestoresApi.ApiRestoresIdGet(ctx, restoreID).Execute()
		api.RequireNoError(t, resp, err)
		state := restore.GetStatus()
		require.Equal(t, "Failed", state, "Restore %q is in state %q.", restoreID, state)
	})
}

func (c *ControlPlane) RetryRestore(ctx context.Context, t tests.T, restoreID string) *pds.ModelsRestore {
	restore, resp, err := c.PDS.RestoresApi.ApiRestoresIdRetryPost(ctx, restoreID).Execute()
	api.RequireNoError(t, resp, err)
	require.NotNil(t, restore)
	return restore
}
