package controlplane

import (
	"context"
	"testing"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (c *ControlPlane) MustCreateRestore(ctx context.Context, t *testing.T, backupJobID string, name string, nsID, deploymentTargetID *string) *pds.ModelsRestore {
	requestBody := pds.RequestsCreateRestoreRequest{
		DeploymentTargetId: deploymentTargetID,
		Name:               &name,
		NamespaceId:        nsID,
	}
	restore, resp, err := c.PDS.RestoresApi.ApiBackupJobsIdRestorePost(ctx, backupJobID).Body(requestBody).Execute()
	api.RequireNoError(t, resp, err)
	return restore
}

func (c *ControlPlane) MustWaitForRestoreSuccessful(ctx context.Context, t tests.T, restoreID string) {
	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		restore, resp, err := c.PDS.RestoresApi.ApiRestoresIdGet(ctx, restoreID).Execute()
		api.RequireNoError(t, resp, err)
		state := restore.GetStatus()
		require.Equal(t, "Successful", state, "Restore %q is in state %q.", restoreID, state)
	})
}
