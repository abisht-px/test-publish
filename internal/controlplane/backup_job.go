package controlplane

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/require"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

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

type ProjectsIdBackupJobsGetRequestOptions func(*pds.ApiApiProjectsIdBackupJobsGetRequest)

func WithListBackupJobsInProjectBackupID(backupID string) ProjectsIdBackupJobsGetRequestOptions {
	return func(r *pds.ApiApiProjectsIdBackupJobsGetRequest) {
		r.BackupId(backupID)
	}
}

func (c *ControlPlane) ListBackupJobsInProject(ctx context.Context, projectID string, opts ...ProjectsIdBackupJobsGetRequestOptions) ([]pds.ModelsBackupJob, *http.Response, error) {
	req := c.PDS.BackupJobsApi.ApiProjectsIdBackupJobsGet(ctx, projectID)

	for _, o := range opts {
		o(&req)
	}

	backupJobList, resp, err := req.Execute()
	if err != nil {
		return nil, resp, err
	}

	return backupJobList.GetData(), resp, err
}

func (c *ControlPlane) MustListBackupJobsInProject(ctx context.Context, t tests.T, projectID string, opts ...ProjectsIdBackupJobsGetRequestOptions) []pds.ModelsBackupJob {
	backupJobs, resp, err := c.ListBackupJobsInProject(ctx, projectID, opts...)
	api.RequireNoError(t, resp, err)
	require.NotEmpty(t, backupJobs)
	return backupJobs
}
