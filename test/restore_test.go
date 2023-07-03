package test

import (
	"fmt"
	"net/http"

	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (s *PDSTestSuite) TestRestore_MissingPXCloudCredentials() {
	if *skipBackups {
		s.T().Skip("Backup tests skipped.")
	}

	// Given
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: "14.6",
		NodeCount:       1,
	}

	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), deploymentID)
	})
	s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForStatefulSetReady(s.ctx, s.T(), deploymentID)
	pdsDeployment, resp, err := s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := s.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(s.ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()
	restoreName := generateRandomName("restore")

	// Setup backup creds
	name := generateRandomName("backup-creds")
	backupTargetConfig := s.config.backupTarget
	s3Creds := backupTargetConfig.credentials.S3
	backupCredentials := s.controlPlane.MustCreateS3BackupCredentials(s.ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target
	backupTarget := s.controlPlane.MustCreateS3BackupTarget(s.ctx, s.T(), backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.controlPlane.MustEnsureBackupTargetCreatedInTC(s.ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackupTarget(s.ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup
	backup := s.controlPlane.MustCreateBackup(s.ctx, s.T(), deploymentID, backupTarget.GetId())
	s.crossCluster.MustEnsureBackupSuccessful(s.ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackup(s.ctx, s.T(), backup.GetId(), false) })

	// When
	pdsBackup, err := s.targetCluster.GetPDSBackup(s.ctx, namespace, backup.GetClusterResourceName())
	s.Require().NoError(err)
	pxCloudCredential, err := s.targetCluster.FindCloudCredentialByName(s.ctx, pdsBackup.Spec.CloudCredentialName)
	s.Require().NoError(err)
	s.Require().NotNil(pxCloudCredential)
	err = s.targetCluster.DeletePXCloudCredential(s.ctx, pxCloudCredential.ID)
	s.Require().NoError(err)
	s.crossCluster.MustCreateRestore(s.ctx, s.T(), namespace, backup.GetClusterResourceName(), restoreName)
	s.T().Cleanup(func() {
		err := s.targetCluster.DeletePDSRestore(s.ctx, namespace, restoreName)
		s.Require().NoError(err)
	})

	// Then
	// Wait for the restore to fail.
	wait.For(s.T(), wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		pdsRestore, err := s.targetCluster.GetPDSRestore(s.ctx, namespace, restoreName)
		require.NoErrorf(t, err, "Getting restore %s/%s for deployment %s from target cluster.", namespace, restoreName, deploymentID)
		require.Equal(t,
			backupsv1.RestoreStatusFailed,
			pdsRestore.Status.CompletionStatus,
			"Restore %s for the deployment %s status must be failed.",
			restoreName,
			deploymentID,
		)
		require.Equal(t,
			backupsv1.PXCloudCredentialsNotFound,
			pdsRestore.Status.ErrorCode,
			"Expected error code PXCloudCredentialsNotFound for Restore %s for the deployment %s",
			restoreName,
			deploymentID,
		)
	})
}

func (s *PDSTestSuite) TestRestoreDelete_CP() {
	if *skipBackups {
		s.T().Skip("Backup tests skipped.")
	}
	// Create a new deployment.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.1.2",
		NodeCount:       1,
		CRDNamePlural:   "cassandras",
	}

	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), deploymentID)
	})
	s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForStatefulSetReady(s.ctx, s.T(), deploymentID)
	pdsDeployment, resp, err := s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := s.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(s.ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()

	// Setup backup creds
	name := generateRandomName("backup-creds")
	backupTargetConfig := s.config.backupTarget
	s3Creds := backupTargetConfig.credentials.S3
	backupCredentials := s.controlPlane.MustCreateS3BackupCredentials(s.ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target
	backupTarget := s.controlPlane.MustCreateS3BackupTarget(s.ctx, s.T(), backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.controlPlane.MustEnsureBackupTargetCreatedInTC(s.ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackupTarget(s.ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup
	backup := s.controlPlane.MustCreateBackup(s.ctx, s.T(), deploymentID, backupTarget.GetId())
	s.crossCluster.MustEnsureBackupSuccessful(s.ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackup(s.ctx, s.T(), backup.GetId(), false) })

	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	backupJob, err := s.targetCluster.GetPDSBackupJob(s.ctx, namespace, backupJobName)
	s.Require().NoError(err)
	backupJobId, err := getBackupJobID(backupJob)
	s.Require().NoError(err)
	backupJobCP := s.controlPlane.MustGetBackupJob(s.ctx, s.T(), backupJobId)
	s.Require().Equal(backupJobCP.GetId(), backupJobId)
	s.Require().Equal(backupJobCP.GetBackupId(), backup.GetId())
	//Create Restore
	restoreName := generateRandomName("restore")
	restore := s.controlPlane.MustCreateRestore(s.ctx, s.T(), backupJobId, restoreName, backupJobCP.NamespaceId, backupJobCP.DeploymentTargetId)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeploymentIfExists(s.ctx, s.T(), *restore.DeploymentId)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), *restore.DeploymentId)
	})

	s.controlPlane.MustWaitForRestoreSuccessful(s.ctx, s.T(), *restore.Id)
	s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, s.T(), *restore.DeploymentId)
	s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, s.T(), *restore.DeploymentId)
	s.crossCluster.MustWaitForStatefulSetReady(s.ctx, s.T(), *restore.DeploymentId)
	s.controlPlane.MustWaitForDeploymentAvailable(s.ctx, s.T(), *restore.DeploymentId)

	pdsDeployment, resp, err = s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, *restore.DeploymentId).Execute()
	api.RequireNoError(s.T(), resp, err)

	customResourceName := *pdsDeployment.ClusterResourceName

	// delete restore deployment
	s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), *restore.DeploymentId)
	s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), *restore.DeploymentId)

	_, resp, err = s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, *restore.DeploymentId).Execute()
	require.EqualErrorf(s.T(), err, "404 Not Found", "Expected an error response on getting deployment %s.", deploymentID)
	require.Equalf(s.T(), http.StatusNotFound, resp.StatusCode, "Deployment %s is not removed.", deploymentID)

	// then verify CR is deleted from TC
	cr, err := s.targetCluster.GetPDSDeployment(s.ctx, namespace, deployment.CRDNamePlural, customResourceName)
	require.Nil(s.T(), cr)

	expectedError := fmt.Sprintf("%s.deployments.pds.io %q not found", deployment.CRDNamePlural, customResourceName)
	require.EqualError(s.T(), err, expectedError, "CR is not deleted on TC")
}

func (s *PDSTestSuite) TestRestoreSuccessful() {
	if *skipBackups {
		s.T().Skip("Backup tests skipped.")
	}

	// Given.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.1.2",
		NodeCount:       1,
	}

	// Deploy DS.
	deployment.NamePrefix = fmt.Sprintf("restore-%s-", deployment.ImageVersionString())
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), deploymentID)
	})
	s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForStatefulSetReady(s.ctx, s.T(), deploymentID)
	pdsDeployment, resp, err := s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := s.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(s.ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()
	restoreName := generateRandomName("restore")

	// Setup backup creds.
	name := generateRandomName("pds-creds")
	backupTargetConfig := s.config.backupTarget
	s3Creds := backupTargetConfig.credentials.S3
	backupCredentials := s.controlPlane.MustCreateS3BackupCredentials(s.ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target.
	backupTarget := s.controlPlane.MustCreateS3BackupTarget(s.ctx, s.T(), backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.controlPlane.MustEnsureBackupTargetCreatedInTC(s.ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackupTarget(s.ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup.
	backup := s.controlPlane.MustCreateBackup(s.ctx, s.T(), deploymentID, backupTarget.GetId())
	s.crossCluster.MustEnsureBackupSuccessful(s.ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackup(s.ctx, s.T(), backup.GetId(), false) })

	// Fetch backjob ID.
	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	backupJobTC, err := s.targetCluster.GetPDSBackupJob(s.ctx, namespace, backupJobName)
	s.Require().NoError(err)
	backupJobId, err := getBackupJobID(backupJobTC)
	s.Require().NoError(err)
	backupJobCP := s.controlPlane.MustGetBackupJob(s.ctx, s.T(), backupJobId)

	// When.
	restore := s.controlPlane.MustCreateRestore(s.ctx, s.T(), backupJobId, restoreName, backupJobCP.NamespaceId, backupJobCP.DeploymentTargetId)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), *restore.DeploymentId)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), *restore.DeploymentId)
	})

	// Then.
	s.controlPlane.MustWaitForRestoreSuccessful(s.ctx, s.T(), *restore.Id)
	s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, s.T(), *restore.DeploymentId)
	s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, s.T(), *restore.DeploymentId)
	s.crossCluster.MustWaitForStatefulSetReady(s.ctx, s.T(), *restore.DeploymentId)
	s.controlPlane.MustWaitForDeploymentAvailable(s.ctx, s.T(), *restore.DeploymentId)
}
