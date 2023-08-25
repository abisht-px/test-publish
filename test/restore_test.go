package test

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/portworx/pds-integration-test/internal/controlplane"

	"github.com/stretchr/testify/require"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	postgresImageVersionTag = "14.8"
)

var (
	skipRestore = flag.Bool("skip-restore", false, "Skip tests related to restore.")
)

func (s *PDSTestSuite) TestRestore_MissingPXCloudCredentials() {
	if *skipBackups {
		s.T().Skip("Backup tests skipped.")
	}

	// Given
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: postgresImageVersionTag,
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
	// Create Restore
	restoreName := generateRandomName("restore")
	restore := s.controlPlane.MustCreateRestore(s.ctx, s.T(), backupJobId, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)
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
	restore := s.controlPlane.MustCreateRestore(s.ctx, s.T(), backupJobId, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)
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

func (s *PDSTestSuite) TestRestore_TargetClusterNotSupported() {
	if *skipRestore {
		s.T().Skip("Restore tests skipped.")
	}

	var (
		backupCredentialsID string
		backupTargetID      string
		deploymentID        string
		backupJobID         string
		deploymentTargetID  string

		backupModel            *pds.ModelsBackup
		restoreName            = generateRandomName("restore")
		postgresDeploymentSpec = api.ShortDeploymentSpec{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: postgresImageVersionTag,
			NodeCount:       1,
		}

		namespaceID  = s.controlPlane.TestPDSNamespaceID
		projectID    = s.controlPlane.TestPDSProjectID
		backupTarget = s.config.backupTarget
	)

	s.Run("Get Deployment Target ID", func() {
		s.T().Helper()

		deploymentTarget := s.controlPlane.MustGetDeploymentTarget(s.ctx, s.T())
		deploymentTargetID = deploymentTarget.GetId()
	})

	postgresDeploymentSpec.NamePrefix = postgresDeploymentSpec.ImageVersionString()

	s.Run("Deploy Postgres", func() {
		deploymentID = s.controlPlane.MustDeployDeploymentSpec(s.ctx, s.T(), &postgresDeploymentSpec)
	})

	s.T().Cleanup(func() {
		if deploymentID != "" {
			s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), deploymentID)
			s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), deploymentID)
		}
	})

	s.Run("Wait for Postgres Deployment to be Healthy and Ready", func() {
		s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, s.T(), deploymentID)
		s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, s.T(), deploymentID)
		s.crossCluster.MustWaitForStatefulSetReady(s.ctx, s.T(), deploymentID)
	})

	s.Run("Setup backup creds", func() {
		backupCredentials := s.controlPlane.MustCreateS3BackupCredentials(
			s.ctx, s.T(),
			backupTarget.credentials.S3,
			generateRandomName("backup-creds"),
		)

		backupCredentialsID = backupCredentials.GetId()
	})

	s.T().Cleanup(func() {
		s.controlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), backupCredentialsID)
	})

	s.Run("Setup backup target", func() {
		backupTarget := s.controlPlane.MustCreateS3BackupTarget(
			s.ctx, s.T(),
			backupCredentialsID,
			backupTarget.bucket,
			backupTarget.region,
		)

		backupTargetID = backupTarget.GetId()

		s.controlPlane.MustEnsureBackupTargetCreatedInTC(s.ctx, s.T(), backupTargetID)
	})

	s.T().Cleanup(func() {
		s.controlPlane.MustDeleteBackupTarget(s.ctx, s.T(), backupTargetID)
	})

	s.Run("Take Adhoc backup", func() {
		backupModel = s.controlPlane.MustCreateBackup(s.ctx, s.T(), deploymentID, backupTargetID)
		s.crossCluster.MustEnsureBackupSuccessful(s.ctx, s.T(), deploymentID, backupModel.GetClusterResourceName())
	})

	s.T().Cleanup(func() {
		s.controlPlane.MustDeleteBackup(s.ctx, s.T(), backupModel.GetId(), false)
	})

	s.Run("Retrieve backup job ID", func() {
		s.T().Helper()

		backupJobs := s.controlPlane.MustListBackupJobsInProject(
			s.ctx, s.T(), projectID,
			controlplane.WithListBackupJobsInProjectBackupID(backupModel.GetId()),
		)

		backupJobID = *backupJobs[0].Id
	})

	s.Run("Restore should not get triggered and error out", func() {
		_, resp, err := s.controlPlane.CreateRestore(
			s.ctx,
			backupJobID,
			restoreName,
			namespaceID,
			deploymentTargetID,
		)

		err = api.ExtractErrorDetails(resp, err)
		require.Error(s.T(), err)
		api.RequireErrorWithStatus(s.T(), resp, err, 422)
		require.Contains(s.T(), err.Error(), "incompatible_restore_capabilities")
	})
}

func (s *PDSTestSuite) TestRestore_EditDeploymentWhenRestoreFailed() {
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

	pdsBackup, err := s.targetCluster.GetPDSBackup(s.ctx, namespace, backup.GetClusterResourceName())
	s.Require().NoError(err)
	pxCloudCredential, err := s.targetCluster.FindCloudCredentialByName(s.ctx, pdsBackup.Spec.CloudCredentialName)
	s.Require().NoError(err)
	s.Require().NotNil(pxCloudCredential)
	err = s.targetCluster.DeletePXCloudCredential(s.ctx, pxCloudCredential.ID)
	s.Require().NoError(err)

	// When.
	restore := s.controlPlane.MustCreateRestore(s.ctx, s.T(), backupJobId, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), *restore.DeploymentId)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), *restore.DeploymentId)
	})

	// Then.
	// Wait for the restore to fail.
	s.controlPlane.MustWaitForRestoreFailed(s.ctx, s.T(), *restore.Id)
	s.controlPlane.FailUpdateDeployment(s.ctx, s.T(), *restore.DeploymentId, &deployment)
}

func (s *PDSTestSuite) TestRestore_RetryRestore() {
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

	pdsBackup, err := s.targetCluster.GetPDSBackup(s.ctx, namespace, backup.GetClusterResourceName())
	s.Require().NoError(err)
	pxCloudCredential, err := s.targetCluster.FindCloudCredentialByName(s.ctx, pdsBackup.Spec.CloudCredentialName)
	s.Require().NoError(err)
	s.Require().NotNil(pxCloudCredential)
	err = s.targetCluster.DeletePXCloudCredential(s.ctx, pxCloudCredential.ID)
	s.Require().NoError(err)

	// When.
	restore := s.controlPlane.MustCreateRestore(s.ctx, s.T(), backupJobId, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)

	// Wait for the restore to fail
	s.controlPlane.MustWaitForRestoreFailed(s.ctx, s.T(), *restore.Id)

	// Recreate the credentials with same name in PXNamespace
	err = s.targetCluster.CreatePXCloudCredentialsForS3(s.ctx, pdsBackup.Spec.CloudCredentialName, backupTargetConfig.bucket, s3Creds)
	s.Require().NoError(err)

	// Then.
	retryRestore := s.controlPlane.RetryRestore(s.ctx, s.T(), *restore.Id, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), *retryRestore.DeploymentId)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), *retryRestore.DeploymentId)
	})
	s.controlPlane.MustWaitForRestoreSuccessful(s.ctx, s.T(), *retryRestore.Id)
	s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, s.T(), *retryRestore.DeploymentId)
	s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, s.T(), *retryRestore.DeploymentId)
	s.crossCluster.MustWaitForStatefulSetReady(s.ctx, s.T(), *retryRestore.DeploymentId)
	s.controlPlane.MustWaitForDeploymentAvailable(s.ctx, s.T(), *retryRestore.DeploymentId)
}

func (s *PDSTestSuite) TestRestore_RetryRestore_WithNewRequestBody() {
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

	pdsBackup, err := s.targetCluster.GetPDSBackup(s.ctx, namespace, backup.GetClusterResourceName())
	s.Require().NoError(err)
	pxCloudCredential, err := s.targetCluster.FindCloudCredentialByName(s.ctx, pdsBackup.Spec.CloudCredentialName)
	s.Require().NoError(err)
	s.Require().NotNil(pxCloudCredential)
	err = s.targetCluster.DeletePXCloudCredential(s.ctx, pxCloudCredential.ID)
	s.Require().NoError(err)

	restoreNewName := generateRandomName("restore-new")
	// When.
	restore := s.controlPlane.MustCreateRestore(s.ctx, s.T(), backupJobId, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)

	// Wait for the restore to fail
	s.controlPlane.MustWaitForRestoreFailed(s.ctx, s.T(), *restore.Id)

	// Recreate the credentials with same name in PXNamespace
	err = s.targetCluster.CreatePXCloudCredentialsForS3(s.ctx, pdsBackup.Spec.CloudCredentialName, backupTargetConfig.bucket, s3Creds)
	s.Require().NoError(err)

	// Then.
	retryRestore := s.controlPlane.RetryRestore(s.ctx, s.T(), *restore.Id, restoreNewName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), *retryRestore.DeploymentId)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), *retryRestore.DeploymentId)
	})
	s.controlPlane.MustWaitForRestoreSuccessful(s.ctx, s.T(), *retryRestore.Id)
	s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, s.T(), *retryRestore.DeploymentId)
	s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, s.T(), *retryRestore.DeploymentId)
	s.crossCluster.MustWaitForStatefulSetReady(s.ctx, s.T(), *retryRestore.DeploymentId)
	s.controlPlane.MustWaitForDeploymentAvailable(s.ctx, s.T(), *retryRestore.DeploymentId)
}
