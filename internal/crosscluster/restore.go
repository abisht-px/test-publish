package crosscluster

import (
	"context"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"

	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
	"github.com/stretchr/testify/require"
)

func (c *CrossClusterHelper) MustCreateRestore(ctx context.Context, t tests.T, deploymentID, backupName, restoreName string) *backupsv1.Restore {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespace, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	pdsBackup, err := c.targetCluster.GetPDSBackup(ctx, namespace.GetName(), backupName)
	require.NoError(t, err)

	snapshotID, err := getBackupSnapshotID(pdsBackup)
	require.NoError(t, err)

	restore, err := c.targetCluster.CreatePDSRestore(ctx, namespace.GetName(), restoreName, pdsBackup.Spec.CloudCredentialName, snapshotID)
	require.NoError(t, err)

	return restore
}

func (c *CrossClusterHelper) MustEnsureRestoreSuccessful(ctx context.Context, t tests.T, deploymentID, restoreName string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()

	// 1. Wait for the restore to finish.
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		pdsRestore, err := c.targetCluster.GetPDSRestore(ctx, namespace, restoreName)
		require.NoErrorf(t, err, "Getting restore %s/%s for deployment %s from target cluster.", namespace, restoreName, deploymentID)
		require.Truef(t, isRestoreFinished(pdsRestore), "Restore %s for the deployment %s did not finish.", restoreName, deploymentID)
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
