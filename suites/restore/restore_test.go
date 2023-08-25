package restore_test

import (
	"fmt"
	"net/http"

	"github.com/portworx/pds-integration-test/suites/framework"

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

func (s *RestoreTestSuite) TestRestore_MissingPXCloudCredentials() {
	// Given
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: postgresImageVersionTag,
		NodeCount:       1,
	}

	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
	deploymentID := controlPlane.MustDeployDeploymentSpec(ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), deploymentID)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)
	})
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), deploymentID)
	pdsDeployment, resp, err := controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()
	restoreName := framework.NewRandomName("restore")

	// Setup backup creds
	name := framework.NewRandomName("backup-creds")
	backupTargetConfig := backupTargetCfg
	s3Creds := backupTargetConfig.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target
	backupTarget := controlPlane.MustCreateS3BackupTarget(
		ctx, s.T(),
		backupCredentials.GetId(),
		backupTargetConfig.Bucket,
		backupTargetConfig.Region,
	)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), false) })

	// When
	pdsBackup, err := targetCluster.GetPDSBackup(ctx, namespace, backup.GetClusterResourceName())
	s.Require().NoError(err)
	pxCloudCredential, err := targetCluster.FindCloudCredentialByName(ctx, pdsBackup.Spec.CloudCredentialName)
	s.Require().NoError(err)
	s.Require().NotNil(pxCloudCredential)
	err = targetCluster.DeletePXCloudCredential(ctx, pxCloudCredential.ID)
	s.Require().NoError(err)
	crossCluster.MustCreateRestore(ctx, s.T(), namespace, backup.GetClusterResourceName(), restoreName)
	s.T().Cleanup(func() {
		err := targetCluster.DeletePDSRestore(ctx, namespace, restoreName)
		s.Require().NoError(err)
	})

	// Then
	// Wait for the restore to fail.
	wait.For(s.T(), wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		pdsRestore, err := targetCluster.GetPDSRestore(ctx, namespace, restoreName)
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

func (s *RestoreTestSuite) TestRestoreDelete_CP() {
	// Create a new deployment.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.1.2",
		NodeCount:       1,
		CRDNamePlural:   "cassandras",
	}

	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
	deploymentID := controlPlane.MustDeployDeploymentSpec(ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), deploymentID)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)
	})
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), deploymentID)
	pdsDeployment, resp, err := controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()

	// Setup backup creds
	name := framework.NewRandomName("backup-creds")
	backupTargetConfig := backupTargetCfg
	s3Creds := backupTargetConfig.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetConfig.Bucket, backupTargetConfig.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), false) })

	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	backupJob, err := targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	s.Require().NoError(err)
	backupJobId, err := getBackupJobID(backupJob)
	s.Require().NoError(err)
	backupJobCP := controlPlane.MustGetBackupJob(ctx, s.T(), backupJobId)
	s.Require().Equal(backupJobCP.GetId(), backupJobId)
	s.Require().Equal(backupJobCP.GetBackupId(), backup.GetId())
	// Create Restore
	restoreName := framework.NewRandomName("restore")
	restore := controlPlane.MustCreateRestore(ctx, s.T(), backupJobId, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeploymentIfExists(ctx, s.T(), *restore.DeploymentId)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), *restore.DeploymentId)
	})

	controlPlane.MustWaitForRestoreSuccessful(ctx, s.T(), *restore.Id)
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), *restore.DeploymentId)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), *restore.DeploymentId)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), *restore.DeploymentId)
	controlPlane.MustWaitForDeploymentAvailable(ctx, s.T(), *restore.DeploymentId)

	pdsDeployment, resp, err = controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, *restore.DeploymentId).Execute()
	api.RequireNoError(s.T(), resp, err)

	customResourceName := *pdsDeployment.ClusterResourceName

	// delete restore deployment
	controlPlane.MustRemoveDeployment(ctx, s.T(), *restore.DeploymentId)
	controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), *restore.DeploymentId)

	_, resp, err = controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, *restore.DeploymentId).Execute()
	require.EqualErrorf(s.T(), err, "404 Not Found", "Expected an error response on getting deployment %s.", deploymentID)
	require.Equalf(s.T(), http.StatusNotFound, resp.StatusCode, "Deployment %s is not removed.", deploymentID)

	// then verify CR is deleted from TC
	cr, err := targetCluster.GetPDSDeployment(ctx, namespace, deployment.CRDNamePlural, customResourceName)
	require.Nil(s.T(), cr)

	expectedError := fmt.Sprintf("%s.deployments.pds.io %q not found", deployment.CRDNamePlural, customResourceName)
	require.EqualError(s.T(), err, expectedError, "CR is not deleted on TC")
}

func (s *RestoreTestSuite) TestRestoreSuccessful() {
	// Given.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.1.2",
		NodeCount:       1,
	}

	// Deploy DS.
	deployment.NamePrefix = fmt.Sprintf("restore-%s-", deployment.ImageVersionString())
	deploymentID := controlPlane.MustDeployDeploymentSpec(ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), deploymentID)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)
	})
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), deploymentID)
	pdsDeployment, resp, err := controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()
	restoreName := framework.NewRandomName("restore")

	// Setup backup creds.
	name := framework.NewRandomName("pds-creds")
	backupTargetConfig := backupTargetCfg
	s3Creds := backupTargetConfig.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target.
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetConfig.Bucket, backupTargetConfig.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup.
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), false) })

	// Fetch backjob ID.
	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	backupJobTC, err := targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	s.Require().NoError(err)
	backupJobId, err := getBackupJobID(backupJobTC)
	s.Require().NoError(err)
	backupJobCP := controlPlane.MustGetBackupJob(ctx, s.T(), backupJobId)

	// When.
	restore := controlPlane.MustCreateRestore(ctx, s.T(), backupJobId, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), *restore.DeploymentId)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), *restore.DeploymentId)
	})

	// Then.
	controlPlane.MustWaitForRestoreSuccessful(ctx, s.T(), *restore.Id)
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), *restore.DeploymentId)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), *restore.DeploymentId)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), *restore.DeploymentId)
	controlPlane.MustWaitForDeploymentAvailable(ctx, s.T(), *restore.DeploymentId)
}

func (s *RestoreTestSuite) TestRestore_TargetClusterNotSupported() {
	if skipExtendedTests {
		s.T().Skip("Skipping extneded test suites")
	}

	var (
		backupCredentialsID string
		backupTargetID      string
		deploymentID        string
		backupJobID         string
		deploymentTargetID  string

		backupModel            *pds.ModelsBackup
		restoreName            = framework.NewRandomName("restore")
		postgresDeploymentSpec = api.ShortDeploymentSpec{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: postgresImageVersionTag,
			NodeCount:       1,
		}

		namespaceID  = controlPlane.TestPDSNamespaceID
		projectID    = controlPlane.TestPDSProjectID
		backupTarget = backupTargetCfg
	)

	s.Run("Get Deployment Target ID", func() {
		s.T().Helper()

		deploymentTarget := controlPlane.MustGetDeploymentTarget(ctx, s.T())
		deploymentTargetID = deploymentTarget.GetId()
	})

	postgresDeploymentSpec.NamePrefix = postgresDeploymentSpec.ImageVersionString()

	s.Run("Deploy Postgres", func() {
		deploymentID = controlPlane.MustDeployDeploymentSpec(ctx, s.T(), &postgresDeploymentSpec)
	})

	s.T().Cleanup(func() {
		if deploymentID != "" {
			controlPlane.MustRemoveDeployment(ctx, s.T(), deploymentID)
			controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)
		}
	})

	s.Run("Wait for Postgres Deployment to be Healthy and Ready", func() {
		controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), deploymentID)
		crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), deploymentID)
		crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), deploymentID)
	})

	s.Run("Setup backup creds", func() {
		backupCredentials := controlPlane.MustCreateS3BackupCredentials(
			ctx, s.T(),
			backupTarget.Credentials.S3,
			framework.NewRandomName("backup-creds"),
		)

		backupCredentialsID = backupCredentials.GetId()
	})

	s.T().Cleanup(func() {
		controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentialsID)
	})

	s.Run("Setup backup target", func() {
		backupTarget := controlPlane.MustCreateS3BackupTarget(
			ctx, s.T(),
			backupCredentialsID,
			backupTarget.Bucket,
			backupTarget.Region,
		)

		backupTargetID = backupTarget.GetId()

		controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTargetID)
	})

	s.T().Cleanup(func() {
		controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTargetID)
	})

	s.Run("Take Adhoc backup", func() {
		backupModel = controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTargetID)
		crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backupModel.GetClusterResourceName())
	})

	s.T().Cleanup(func() {
		controlPlane.MustDeleteBackup(ctx, s.T(), backupModel.GetId(), false)
	})

	s.Run("Retrieve backup job ID", func() {
		s.T().Helper()

		backupJobs := controlPlane.MustListBackupJobsInProject(
			ctx, s.T(), projectID,
			controlplane.WithListBackupJobsInProjectBackupID(backupModel.GetId()),
		)

		backupJobID = *backupJobs[0].Id
	})

	s.Run("Restore should not get triggered and error out", func() {
		_, resp, err := controlPlane.CreateRestore(
			ctx,
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

func (s *RestoreTestSuite) TestRestore_EditDeploymentWhenRestoreFailed() {
	// Given.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.1.2",
		NodeCount:       1,
	}

	// Deploy DS.
	deployment.NamePrefix = fmt.Sprintf("restore-%s-", deployment.ImageVersionString())
	deploymentID := controlPlane.MustDeployDeploymentSpec(ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), deploymentID)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)
	})
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), deploymentID)
	pdsDeployment, resp, err := controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()
	restoreName := framework.NewRandomName("restore")

	// Setup backup creds.
	name := framework.NewRandomName("pds-creds")
	backupTargetConfig := backupTargetCfg
	s3Creds := backupTargetConfig.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target.
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetConfig.Bucket, backupTargetConfig.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup.
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), false) })

	// Fetch backjob ID.
	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	backupJobTC, err := targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	s.Require().NoError(err)
	backupJobId, err := getBackupJobID(backupJobTC)
	s.Require().NoError(err)
	backupJobCP := controlPlane.MustGetBackupJob(ctx, s.T(), backupJobId)

	pdsBackup, err := targetCluster.GetPDSBackup(ctx, namespace, backup.GetClusterResourceName())
	s.Require().NoError(err)
	pxCloudCredential, err := targetCluster.FindCloudCredentialByName(ctx, pdsBackup.Spec.CloudCredentialName)
	s.Require().NoError(err)
	s.Require().NotNil(pxCloudCredential)
	err = targetCluster.DeletePXCloudCredential(ctx, pxCloudCredential.ID)
	s.Require().NoError(err)

	// When.
	restore := controlPlane.MustCreateRestore(ctx, s.T(), backupJobId, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), *restore.DeploymentId)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), *restore.DeploymentId)
	})

	// Then.
	// Wait for the restore to fail.
	controlPlane.MustWaitForRestoreFailed(ctx, s.T(), *restore.Id)
	controlPlane.FailUpdateDeployment(ctx, s.T(), *restore.DeploymentId, &deployment)
}

func (s *RestoreTestSuite) TestRestore_RetryRestore() {
	// Given.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.1.2",
		NodeCount:       1,
	}

	// Deploy DS.
	deployment.NamePrefix = fmt.Sprintf("restore-%s-", deployment.ImageVersionString())
	deploymentID := controlPlane.MustDeployDeploymentSpec(ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), deploymentID)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)
	})
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), deploymentID)
	pdsDeployment, resp, err := controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()
	restoreName := framework.NewRandomName("restore")

	// Setup backup creds.
	name := framework.NewRandomName("pds-creds")
	backupTargetConfig := backupTargetCfg
	s3Creds := backupTargetConfig.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target.
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetConfig.Bucket, backupTargetConfig.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup.
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), false) })

	// Fetch backjob ID.
	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	backupJobTC, err := targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	s.Require().NoError(err)
	backupJobId, err := getBackupJobID(backupJobTC)
	s.Require().NoError(err)
	backupJobCP := controlPlane.MustGetBackupJob(ctx, s.T(), backupJobId)

	pdsBackup, err := targetCluster.GetPDSBackup(ctx, namespace, backup.GetClusterResourceName())
	s.Require().NoError(err)
	pxCloudCredential, err := targetCluster.FindCloudCredentialByName(ctx, pdsBackup.Spec.CloudCredentialName)
	s.Require().NoError(err)
	s.Require().NotNil(pxCloudCredential)
	err = targetCluster.DeletePXCloudCredential(ctx, pxCloudCredential.ID)
	s.Require().NoError(err)

	// When.
	restore := controlPlane.MustCreateRestore(ctx, s.T(), backupJobId, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)

	// Wait for the restore to fail
	controlPlane.MustWaitForRestoreFailed(ctx, s.T(), *restore.Id)

	// Recreate the credentials with same name in PXNamespace
	err = targetCluster.CreatePXCloudCredentialsForS3(ctx, pdsBackup.Spec.CloudCredentialName, backupTargetConfig.Bucket, s3Creds)
	s.Require().NoError(err)

	// Then.
	retryRestore := controlPlane.RetryRestore(ctx, s.T(), *restore.Id, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), *retryRestore.DeploymentId)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), *retryRestore.DeploymentId)
	})
	controlPlane.MustWaitForRestoreSuccessful(ctx, s.T(), *retryRestore.Id)
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), *retryRestore.DeploymentId)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), *retryRestore.DeploymentId)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), *retryRestore.DeploymentId)
	controlPlane.MustWaitForDeploymentAvailable(ctx, s.T(), *retryRestore.DeploymentId)
}

func (s *RestoreTestSuite) TestRestore_RetryRestore_WithNewRequestBody() {
	// Given.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.1.2",
		NodeCount:       1,
	}

	// Deploy DS.
	deployment.NamePrefix = fmt.Sprintf("restore-%s-", deployment.ImageVersionString())
	deploymentID := controlPlane.MustDeployDeploymentSpec(ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), deploymentID)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)
	})
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), deploymentID)
	pdsDeployment, resp, err := controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()
	restoreName := framework.NewRandomName("restore")

	// Setup backup creds.
	name := framework.NewRandomName("pds-creds")
	backupTargetConfig := backupTargetCfg
	s3Creds := backupTargetConfig.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target.
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetConfig.Bucket, backupTargetConfig.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup.
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), false) })

	// Fetch backjob ID.
	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	backupJobTC, err := targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	s.Require().NoError(err)
	backupJobId, err := getBackupJobID(backupJobTC)
	s.Require().NoError(err)
	backupJobCP := controlPlane.MustGetBackupJob(ctx, s.T(), backupJobId)

	pdsBackup, err := targetCluster.GetPDSBackup(ctx, namespace, backup.GetClusterResourceName())
	s.Require().NoError(err)
	pxCloudCredential, err := targetCluster.FindCloudCredentialByName(ctx, pdsBackup.Spec.CloudCredentialName)
	s.Require().NoError(err)
	s.Require().NotNil(pxCloudCredential)
	err = targetCluster.DeletePXCloudCredential(ctx, pxCloudCredential.ID)
	s.Require().NoError(err)

	restoreNewName := framework.NewRandomName("restore-new")
	// When.
	restore := controlPlane.MustCreateRestore(ctx, s.T(), backupJobId, restoreName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)

	// Wait for the restore to fail
	controlPlane.MustWaitForRestoreFailed(ctx, s.T(), *restore.Id)

	// Recreate the credentials with same name in PXNamespace
	err = targetCluster.CreatePXCloudCredentialsForS3(ctx, pdsBackup.Spec.CloudCredentialName, backupTargetConfig.Bucket, s3Creds)
	s.Require().NoError(err)

	// Then.
	retryRestore := controlPlane.RetryRestore(ctx, s.T(), *restore.Id, restoreNewName, *backupJobCP.NamespaceId, *backupJobCP.DeploymentTargetId)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, s.T(), *retryRestore.DeploymentId)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), *retryRestore.DeploymentId)
	})
	controlPlane.MustWaitForRestoreSuccessful(ctx, s.T(), *retryRestore.Id)
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), *retryRestore.DeploymentId)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), *retryRestore.DeploymentId)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), *retryRestore.DeploymentId)
	controlPlane.MustWaitForDeploymentAvailable(ctx, s.T(), *retryRestore.DeploymentId)
}
