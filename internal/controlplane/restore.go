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

func (c *ControlPlane) MustCreateRestore(ctx context.Context, t tests.T, backupJobID, name, nsID string) *pds.ModelsRestore {
	restore, resp, err := c.CreateRestore(ctx, backupJobID, name, nsID, c.testPDSDeploymentTargetID)
	api.RequireNoError(t, resp, err)
	require.NotNil(t, restore)
	return restore
}

func (c *ControlPlane) CreateRestore(ctx context.Context, backupJobID, name, nsID, deploymentTargetID string) (*pds.ModelsRestore, *http.Response, error) {
	createRestoreReq := c.PDS.RestoresApi.ApiBackupJobsIdRestorePost(ctx, backupJobID)
	createRestoreBody := pds.RequestsCreateRestoreRequest{
		DeploymentTargetId: &deploymentTargetID,
		Name:               &name,
		NamespaceId:        &nsID,
	}
	createRestoreReq = createRestoreReq.Body(createRestoreBody)
	restore, resp, err := c.PDS.RestoresApi.ApiBackupJobsIdRestorePostExecute(createRestoreReq)
	return restore, resp, err
}

func (c *ControlPlane) MustWaitForRestoreSuccessful(ctx context.Context, t tests.T, restoreID string) {
	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		restore, resp, err := c.PDS.RestoresApi.ApiRestoresIdGet(ctx, restoreID).Execute()
		api.RequireNoError(t, resp, err)
		state := restore.GetStatus()
		require.Equal(t, "Successful", state, "Restore %q is in state %q.", restoreID, state)
	})
}
