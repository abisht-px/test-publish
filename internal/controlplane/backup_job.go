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

func (c *ControlPlane) MustWaitForBackupJobRemoved(ctx context.Context, t tests.T, backupJobID string) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		backupJob, resp, err := c.PDS.BackupJobsApi.ApiBackupJobsIdGet(ctx, backupJobID).Execute()
		require.EqualErrorf(t, err, "404 Not Found", "Expected an error response on getting backupjob %s.", backupJobID)
		require.Equalf(t, http.StatusNotFound, resp.StatusCode, "Backup job %s is not removed.", backupJobID)
		require.Nil(t, backupJob)
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
		require.GreaterOrEqual(t, expectedBackups, successfulBackupJobs, "Expected at least %v successful backup jobs", expectedBackups)
	})
}
