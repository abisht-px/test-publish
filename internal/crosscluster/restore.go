package crosscluster

import (
	"context"
	"time"

	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"

	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
	"github.com/stretchr/testify/require"
)

func (c *CrossClusterHelper) MustCreateRestore(ctx context.Context, t tests.T, namespace, backupName, restoreName string) *backupsv1.Restore {
	pdsBackup, err := c.targetCluster.GetPDSBackup(ctx, namespace, backupName)
	require.NoError(t, err)

	snapshotID, err := GetBackupSnapshotID(pdsBackup)
	require.NoError(t, err)

	restore, err := c.targetCluster.CreatePDSRestore(ctx, namespace, restoreName, pdsBackup.Spec.CloudCredentialName, snapshotID)
	require.NoError(t, err)

	return restore
}

func (c *CrossClusterHelper) MustEnsureRestoreSuccessful(ctx context.Context, t tests.T, namespace, restoreName string, waitTimeout time.Duration) {

	// 1. Wait for the restore to finish.
	wait.For(t, waitTimeout, wait.RetryInterval, func(t tests.T) {
		pdsRestore, err := c.targetCluster.GetPDSRestore(ctx, namespace, restoreName)
		require.NoErrorf(t, err, "Getting restore %s/%s from target cluster.", namespace, restoreName)
		require.Truef(t, isRestoreFinished(pdsRestore), "Restore %s did not finish.", restoreName)
	})

	// 2. Check the result.
	pdsRestore, err := c.targetCluster.GetPDSRestore(ctx, namespace, restoreName)
	require.NoError(t, err)

	require.True(t, isRestoreSucceeded(pdsRestore))
}

func isRestoreFinished(restore *backupsv1.Restore) bool {
	return isRestoreSucceeded(restore) || isRestoreFailed(restore)
}

func isRestoreSucceeded(restore *backupsv1.Restore) bool {
	return restore.Status.CompletionStatus == backupsv1.RestoreStatusSuccessful
}

func isRestoreFailed(restore *backupsv1.Restore) bool {
	return restore.Status.CompletionStatus == backupsv1.RestoreStatusFailed
}
