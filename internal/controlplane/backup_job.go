package controlplane

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (c *ControlPlane) MustWaitForBackupJobRemoved(ctx context.Context, t tests.T, backupID string, backupJobName string) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		backupJobs, resp, err := c.PDS.BackupJobsApi.ApiBackupsIdJobsGet(ctx, backupID).Execute()
		require.NoError(t, err, "Expected no error response on getting backup jobs for backup %s.", backupID)
		require.NotNilf(t, resp, "Received no response body while getting backup jobs for backup %s.", backupID)
		for _, job := range backupJobs.Data {
			require.Equalf(t, backupJobName, *job.Name, "Backup job %s for backup %s is not removed.", backupJobName, backupID)
		}
	})
}

func (c *ControlPlane) MustDeleteBackupJobByName(ctx context.Context, t tests.T, backupID string, backupJobName string) {
	resp, err := c.PDS.BackupsApi.ApiBackupsIdJobsNameDelete(ctx, backupID, backupJobName).Execute()
	api.RequireNoError(t, resp, err)
}
