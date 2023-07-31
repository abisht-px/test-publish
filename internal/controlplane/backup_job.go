package controlplane

import (
	"context"
	"net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (c *ControlPlane) MustWaitForBackupJobsRemoved(ctx context.Context, t tests.T, backupID string, backupJobName string) {
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

func (c *ControlPlane) MustDeleteBackupJobByID(ctx context.Context, t tests.T, backupJobID string) {
	resp, err := c.PDS.BackupJobsApi.ApiBackupJobsIdDelete(ctx, backupJobID).Execute()
	api.RequireNoError(t, resp, err)
}

type ProjectsIdBackupJobsGetRequestOptions func(pds.ApiApiProjectsIdBackupJobsGetRequest) pds.ApiApiProjectsIdBackupJobsGetRequest

func WithListBackupJobsInProjectBackupID(backupID string) ProjectsIdBackupJobsGetRequestOptions {
	return func(r pds.ApiApiProjectsIdBackupJobsGetRequest) pds.ApiApiProjectsIdBackupJobsGetRequest {
		return r.BackupId(backupID)
	}
}

func (c *ControlPlane) ListBackupJobsInProject(ctx context.Context, projectID string, opts ...ProjectsIdBackupJobsGetRequestOptions) ([]pds.ModelsBackupJob, *http.Response, error) {
	req := c.PDS.BackupJobsApi.ApiProjectsIdBackupJobsGet(ctx, projectID)

	for _, o := range opts {
		req = o(req)
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

func (c *ControlPlane) MustEnsureNBackupJobsSuccessFromSchedule(ctx context.Context, t tests.T, projectID, backupID string, expectedBackups int) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		backupJobs := c.MustListBackupJobsInProject(ctx, t, projectID, WithListBackupJobsInProjectBackupID(backupID))
		successfulBackupJobs := 0
		for _, backupJob := range backupJobs {
			if backupJob.HasCompletionStatus() && *backupJob.CompletionStatus == string(backupsv1.BackupJobSucceeded) {
				successfulBackupJobs++
			}
		}
		require.GreaterOrEqual(t, successfulBackupJobs, expectedBackups, "Expected at least %v successful backup jobs", expectedBackups)
	})
}
