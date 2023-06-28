package test

import (
	"fmt"

	"github.com/stretchr/testify/require"

	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/random"
)

func (s *PDSTestSuite) TestBackupJobReporting_CP() {
	// Create a new deployment.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.1.2",
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
	backupJob, _ := s.targetCluster.GetPDSBackupJob(s.ctx, namespace, backupJobName)
	backupJobId, _ := getBackupJobID(backupJob)
	backupJob1 := s.controlPlane.MustGetBackupJob(s.ctx, s.T(), backupJobId)
	s.Require().Equal(backupJob1.GetId(), backupJobId)
	s.Require().Equal(backupJob1.GetBackupId(), backup.GetId())
}

func getBackupJobID(backupJob *backupsv1.BackupJob) (string, error) {
	backupJobID := string(backupJob.GetUID())
	if backupJobID == "" {
		return "", fmt.Errorf("backupJob id is empty")
	}
	return backupJobID, nil
}

func (s *PDSTestSuite) TestBackupJobDeletion_FromCP_WithAdhocBackup() {
	if *skipBackups {
		s.T().Skip("Backup tests skipped.")
	}
	// Given
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.1.2",
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

	// When
	s.controlPlane.MustDeleteBackup(s.ctx, s.T(), backup.GetId(), false)
	s.controlPlane.MustWaitForBackupRemoved(s.ctx, s.T(), backup.GetId())

	// Then
	// Check Backup cr deleted
	_, err = s.targetCluster.GetPDSBackup(s.ctx, namespace, backup.GetClusterResourceName())
	expectedError := fmt.Sprintf("backups.backups.pds.io %q not found", backup.GetClusterResourceName())
	require.EqualError(s.T(), err, expectedError)
	// Check BackupJob cr deleted
	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	_, err = s.targetCluster.GetPDSBackupJob(s.ctx, namespace, backupJobName)
	expectedError = fmt.Sprintf("backupjobs.backups.pds.io %q not found", backupJobName)
	require.EqualError(s.T(), err, expectedError)
	// Check VolumeSnapshot cr deleted
	volumeSnapshotName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	_, err = s.targetCluster.GetVolumeSnapshot(s.ctx, namespace, volumeSnapshotName)
	expectedError = fmt.Sprintf("volumesnapshots.volumesnapshot.external-storage.k8s.io %q not found", volumeSnapshotName)
	require.EqualError(s.T(), err, expectedError)
}

func (s *PDSTestSuite) TestBackupJobDeletion_FromCP_WithBackupSchedule() {
	if *skipBackups {
		s.T().Skip("Backup tests skipped.")
	}
	// Given
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.1.2",
		NodeCount:       1,
	}
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
	// Setup backup policy
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	backupPolicyName := fmt.Sprintf("integration-test-%s", nameSuffix)
	schedule := "*/5 * * * *"
	var retention int32 = 4
	backupPolicy := s.controlPlane.MustCreateBackupPolicy(s.ctx, s.T(), &backupPolicyName, &schedule, &retention)
	s.T().Cleanup(func() {
		_, _ = s.controlPlane.DeleteBackupPolicy(s.ctx, backupPolicy.GetId())
	})
	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
	deployment.BackupPolicyname = *backupPolicy.Name
	deployment.BackupTargetName = *backupTarget.Name
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
	// Wait for at least one backup to complete
	backup := s.controlPlane.MustWaitForScheduleBackup(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustEnsureBackupSuccessful(s.ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() {
		s.controlPlane.MustDeleteBackup(s.ctx, s.T(), backup.GetId(), false)
		s.controlPlane.MustWaitForBackupRemoved(s.ctx, s.T(), backup.GetId())
	})

	// When
	pdsBackup, err := s.targetCluster.GetPDSBackup(s.ctx, namespace, backup.GetClusterResourceName())
	require.NoError(s.T(), err)
	backupJobName := pdsBackup.Status.BackupJobs[0].Name
	s.controlPlane.MustDeleteBackupJobByName(s.ctx, s.T(), backup.GetId(), backupJobName)

	// Then
	// Check Backup cr deleted
	s.controlPlane.MustWaitForBackupJobRemoved(s.ctx, s.T(), backup.GetId(), backupJobName)
	// Check BackupJob cr deleted
	_, err = s.targetCluster.GetPDSBackupJob(s.ctx, namespace, backupJobName)
	expectedError := fmt.Sprintf("backupjobs.backups.pds.io %q not found", backupJobName)
	require.EqualError(s.T(), err, expectedError)
	// Check VolumeSnapshot cr deleted
	volumeSnapshotName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	_, err = s.targetCluster.GetVolumeSnapshot(s.ctx, namespace, volumeSnapshotName)
	expectedError = fmt.Sprintf("volumesnapshots.volumesnapshot.external-storage.k8s.io %q not found", volumeSnapshotName)
	require.EqualError(s.T(), err, expectedError)
}