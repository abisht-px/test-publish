package controlplane

import (
	"context"
	"net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/pointer"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (c *ControlPlane) MustCreateBackup(ctx context.Context, t tests.T, deploymentID, backupTargetID string) *pds.ModelsBackup {
	requestBody := pds.ControllersCreateDeploymentBackup{
		BackupLevel:    pointer.String("snapshot"),
		BackupTargetId: pointer.String(backupTargetID),
		BackupType:     pointer.String("adhoc"),
	}
	backup, resp, err := c.PDS.BackupsApi.ApiDeploymentsIdBackupsPost(ctx, deploymentID).Body(requestBody).Execute()
	api.RequireNoError(t, resp, err)

	return backup
}

func (c *ControlPlane) MustDeleteBackup(ctx context.Context, t tests.T, backupID string, localOnly bool) {
	resp, err := c.PDS.BackupsApi.ApiBackupsIdDelete(ctx, backupID).LocalOnly(localOnly).Execute()
	api.RequireNoError(t, resp, err)
}

func (c *ControlPlane) MustGetBackupJob(ctx context.Context, t tests.T, backupJobID string) *pds.ModelsBackupJob {
	backupJob, resp, err := c.PDS.BackupJobsApi.ApiBackupJobsIdGet(ctx, backupJobID).Execute()
	api.RequireNoError(t, resp, err)
	require.NotNil(t, backupJob)
	return backupJob
}

func (c *ControlPlane) MustWaitForBackupRemoved(ctx context.Context, t tests.T, backupID string) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		_, resp, err := c.PDS.BackupsApi.ApiBackupsIdGet(ctx, backupID).Execute()
		require.Errorf(t, err, "Expected an error response on getting backup %s.", backupID)
		require.NotNilf(t, resp, "Received no response body while getting backup %s.", backupID)
		require.Equalf(t, http.StatusNotFound, resp.StatusCode, "Backup %s is not removed.", backupID)
	})
}

func (c *ControlPlane) MustWaitForScheduleBackup(ctx context.Context, t tests.T, deploymentID string) pds.ModelsBackup {
	var err error
	var resp *http.Response
	backups := &pds.ModelsPaginatedResultModelsBackup{}
	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		backups, resp, err = c.PDS.BackupsApi.ApiDeploymentsIdBackupsGet(ctx, deploymentID).SortBy("created_at").Execute()
		api.RequireNoErrorf(t, resp, err, "getting backup for deployment %s.", deploymentID)
		require.NotEmpty(t, backups.GetData(), "Expected atleast one backup for deployment %s.", deploymentID)
	})
	return backups.GetData()[len(backups.GetData())-1]
}
