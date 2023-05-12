package crosscluster

import (
	"context"
	"fmt"

	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (c *CrossClusterHelper) MustEnsureBackupSuccessful(ctx context.Context, t tests.T, deploymentID, backupName string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()

	// 1. Wait for the backup to finish.
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		pdsBackup, err := c.targetCluster.GetPDSBackup(ctx, namespace, backupName)
		require.NoErrorf(t, err, "Getting backup %s/%s for deployment %s from target cluster.", namespace, backupName, deploymentID)
		require.Truef(t, isBackupFinished(pdsBackup), "Backup %s for the deployment %s did not finish.", backupName, deploymentID)
	})

	// 2. Check the result.
	pdsBackup, err := c.targetCluster.GetPDSBackup(ctx, namespace, backupName)
	require.NoError(t, err)

	if isBackupFailed(pdsBackup) {
		// Backup failed.
		backupJobs := pdsBackup.Status.BackupJobs
		var backupJobName string
		if len(backupJobs) > 0 {
			backupJobName = backupJobs[0].Name
		}
		logs, err := c.targetCluster.GetJobLogs(ctx, namespace, backupJobName, c.startTime)
		if err != nil {
			require.Fail(t, fmt.Sprintf("Backup '%s' failed.", backupName))
		} else {
			require.Fail(t, fmt.Sprintf("Backup job '%s' failed. See job logs for more details:", backupJobName), logs)
		}
	}
	require.True(t, isBackupSucceeded(pdsBackup))
}

func getBackupSnapshotID(backup *backupsv1.Backup) (string, error) {
	if !isBackupSucceeded(backup) {
		return "", fmt.Errorf("no succeded backups")
	}
	var snapshotID string
	if len(backup.Status.BackupJobs) > 0 {
		snapshotID = backup.Status.BackupJobs[0].CloudSnapID
	}
	if snapshotID == "" {
		return "", fmt.Errorf("snapshot id is empty")
	}
	return snapshotID, nil
}

func isBackupFinished(backup *backupsv1.Backup) bool {
	return isBackupSucceeded(backup) || isBackupFailed(backup)
}

func isBackupSucceeded(backup *backupsv1.Backup) bool {
	return backup.Status.Succeeded > 0
}

func isBackupFailed(backup *backupsv1.Backup) bool {
	return backup.Status.Failed > 0
}
