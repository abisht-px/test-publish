package backupjob_test

import (
	"fmt"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"

	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/internal/random"
	"github.com/portworx/pds-integration-test/suites/framework"
)

const target = "pds-operator-target-controller-manager"
const fakeEndPoint = "https://ci.pds-dev.io/api"

func (s *BackupJobTestSuite) TestBackupJobReporting_CP() {
	// Create a new deployment.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
		NodeCount:       1,
	}

	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backupjob-%s-", deployment.ImageVersionString())
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
	name := framework.NewRandomName("backupjob-creds")
	s3Creds := backupTargetCfg.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetCfg.Bucket, backupTargetCfg.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), false) })

	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	backupJob, _ := targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	backupJobId, _ := getBackupJobID(backupJob)
	backupJob1 := controlPlane.MustGetBackupJob(ctx, s.T(), backupJobId)
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

func (s *BackupJobTestSuite) TestBackupJobDeletionFromTC_WithDisconnected_TC() {
	// Create a new deployment.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
		NodeCount:       1,
	}

	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backupjob-%s-", deployment.ImageVersionString())
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
	name := framework.NewRandomName("backupjob-creds")
	s3Creds := backupTargetCfg.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })

	// Setup backup target
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetCfg.Bucket, backupTargetCfg.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), true) })

	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	backupJob, _ := targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	backupJobId, _ := getBackupJobID(backupJob)
	backupJob1 := controlPlane.MustGetBackupJob(ctx, s.T(), backupJobId)
	s.Require().Equal(backupJob1.GetId(), backupJobId)
	s.Require().Equal(backupJob1.GetBackupId(), backup.GetId())

	// dissconnect Target cluster
	s.UpdateAPIEndPoint(fakeEndPoint, target)
	s.T().Cleanup(func() { s.UpdateAPIEndPoint(targetCluster.PDSChartConfig.ControlPlaneAPI, target) })
	// Delete backup
	targetCluster.MustDeleteBackupCustomResource(ctx, s.T(), namespace, backup.GetClusterResourceName())
	_, err = targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	expectedError := fmt.Sprintf("backupjobs.backups.pds.io %q not found", backupJobName)
	require.EqualError(s.T(), err, expectedError)
}

func (s *BackupJobTestSuite) UpdateAPIEndPoint(EndPoint string, name string) {
	patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"env":[{"name":"PDS_API_ENDPOINT","value": %q}],"name":"manager"}]}}}}`, EndPoint)
	_, err := targetCluster.PatchDeployment(ctx, targetcluster.PDSChartNamespace, name, []byte(patch))
	if err != nil {
		panic(err)
	}
}

func (s *BackupJobTestSuite) TestBackupJobDeletion_FromCP_WithAdhocBackup() {
	// Given
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
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
	// Setup backup creds
	name := framework.NewRandomName("backupjob-creds")
	s3Creds := backupTargetCfg.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })
	// Setup backup target
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetCfg.Bucket, backupTargetCfg.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })
	// Take Adhoc backup
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())

	// When
	controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), false)
	controlPlane.MustWaitForBackupRemoved(ctx, s.T(), backup.GetId())

	// Then
	// Check Backup cr deleted
	_, err = targetCluster.GetPDSBackup(ctx, namespace, backup.GetClusterResourceName())
	expectedError := fmt.Sprintf("backups.backups.pds.io %q not found", backup.GetClusterResourceName())
	require.EqualError(s.T(), err, expectedError)
	// Check BackupJob cr deleted
	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	_, err = targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	expectedError = fmt.Sprintf("backupjobs.backups.pds.io %q not found", backupJobName)
	require.EqualError(s.T(), err, expectedError)
	// Check VolumeSnapshot cr deleted
	volumeSnapshotName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	_, err = targetCluster.GetVolumeSnapshot(ctx, namespace, volumeSnapshotName)
	expectedError = fmt.Sprintf("volumesnapshots.volumesnapshot.external-storage.k8s.io %q not found", volumeSnapshotName)
	require.EqualError(s.T(), err, expectedError)
}

func (s *BackupJobTestSuite) TestBackupJobDeletion_FromCP_WithBackupSchedule() {
	// Given
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
		NodeCount:       1,
	}
	// Setup backup creds
	name := framework.NewRandomName("backupjob-creds")
	s3Creds := backupTargetCfg.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })
	// Setup backup target
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetCfg.Bucket, backupTargetCfg.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })
	// Setup backup policy
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	backupPolicyName := fmt.Sprintf("integration-test-%s", nameSuffix)
	schedule := "*/5 * * * *"
	var retention int32 = 4
	backupPolicy := controlPlane.MustCreateBackupPolicy(ctx, s.T(), &backupPolicyName, &schedule, &retention)
	s.T().Cleanup(func() {
		_, _ = controlPlane.DeleteBackupPolicy(ctx, backupPolicy.GetId())
	})
	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
	deployment.BackupPolicyname = *backupPolicy.Name
	deployment.BackupTargetName = *backupTarget.Name
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
	// Wait for at least one backup to complete
	backup := controlPlane.MustWaitForScheduleBackup(ctx, s.T(), deploymentID)
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() {
		controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), false)
		controlPlane.MustWaitForBackupRemoved(ctx, s.T(), backup.GetId())
	})

	// When
	backupJobs := controlPlane.MustListBackupJobsInProject(ctx, s.T(), controlPlane.TestPDSProjectID, controlplane.WithListBackupJobsInProjectBackupID(backup.GetId()))
	controlPlane.MustDeleteBackupJobByName(ctx, s.T(), backup.GetId(), *backupJobs[0].Name)

	// Then
	// Check Backup cr deleted
	controlPlane.MustWaitForBackupJobRemoved(ctx, s.T(), backupJobs[0].GetId())
	// Check BackupJob cr deleted
	_, err = targetCluster.GetPDSBackupJob(ctx, namespace, *backupJobs[0].Name)
	expectedError := fmt.Sprintf("backupjobs.backups.pds.io %q not found", *backupJobs[0].Name)
	require.EqualError(s.T(), err, expectedError)
	// Check VolumeSnapshot cr deleted
	volumeSnapshotName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	_, err = targetCluster.GetVolumeSnapshot(ctx, namespace, volumeSnapshotName)
	expectedError = fmt.Sprintf("volumesnapshots.volumesnapshot.external-storage.k8s.io %q not found", volumeSnapshotName)
	require.EqualError(s.T(), err, expectedError)
}

func (s *BackupJobTestSuite) TestBackupJobDeletion_FromTC() {
	// Given
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
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
	// Setup backup creds
	name := framework.NewRandomName("backupjob-creds")
	s3Creds := backupTargetCfg.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })
	// Setup backup target
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetCfg.Bucket, backupTargetCfg.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })
	// Take Adhoc backup
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() {
		controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), true)
	})
	backupJobs := controlPlane.MustListBackupJobsInProject(ctx, s.T(), controlPlane.TestPDSProjectID, controlplane.WithListBackupJobsInProjectBackupID(backup.GetId()))

	// When
	targetCluster.MustDeleteBackupCustomResource(ctx, s.T(), namespace, backup.GetClusterResourceName())

	// Then
	// Check BackupJob in CP deleted
	controlPlane.MustWaitForBackupJobRemoved(ctx, s.T(), *backupJobs[0].Id)
	// Check BackupJob cr deleted
	_, err = targetCluster.GetPDSBackupJob(ctx, namespace, *backupJobs[0].Name)
	expectedError := fmt.Sprintf("backupjobs.backups.pds.io %q not found", *backupJobs[0].Name)
	require.EqualError(s.T(), err, expectedError)
	// Check VolumeSnapshot cr deleted
	volumeSnapshotName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	_, err = targetCluster.GetVolumeSnapshot(ctx, namespace, volumeSnapshotName)
	expectedError = fmt.Sprintf("volumesnapshots.volumesnapshot.external-storage.k8s.io %q not found", volumeSnapshotName)
	require.EqualError(s.T(), err, expectedError)
}

func (s *BackupJobTestSuite) TestBackupJobDeletionFromCP_WithDisconnected_TC() {
	// Create a new deployment.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
		NodeCount:       1,
	}

	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backupjob-%s-", deployment.ImageVersionString())
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
	name := framework.NewRandomName("backupjob-creds")
	s3Creds := backupTargetCfg.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, name)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId()) })
	// Setup backup target
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetCfg.Bucket, backupTargetCfg.Region)
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId()) })

	// Take Adhoc backup
	backup := controlPlane.MustCreateBackup(ctx, s.T(), deploymentID, backupTarget.GetId())
	crossCluster.MustEnsureBackupSuccessful(ctx, s.T(), deploymentID, backup.GetClusterResourceName())
	s.T().Cleanup(func() { controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), true) })

	backupJobName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	backupJob, _ := targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	backupJobId, _ := getBackupJobID(backupJob)
	backupJob1 := controlPlane.MustGetBackupJob(ctx, s.T(), backupJobId)
	s.Require().Equal(backupJob1.GetId(), backupJobId)
	s.Require().Equal(backupJob1.GetBackupId(), backup.GetId())

	// dissconnect Target cluster
	deployment1, _ := targetCluster.GetDeployment(ctx, targetcluster.PDSChartNamespace, "pds-teleport")
	actualReplicas := deployment1.Spec.Replicas
	s.DisconnectTC(0, deployment1)
	controlPlane.DisconnectTestDeploymentTarget(ctx, s.T())

	// Delete backup, but it should not success
	controlPlane.MustDeleteBackupJobWithDisconnectTC(ctx, s.T(), backupJobId)
	backupJobTC := controlPlane.MustGetBackupJob(ctx, s.T(), backupJobId)
	s.Require().Equal(backupJobTC.GetId(), backupJobId)
	// connect TC
	deployment1, _ = targetCluster.GetDeployment(ctx, targetcluster.PDSChartNamespace, "pds-teleport")
	s.DisconnectTC(*actualReplicas, deployment1)
	controlPlane.MustWaitForDeploymentTarget(ctx, s.T(), framework.DeploymentTargetName)

	// Delete backup
	targetCluster.MustDeleteBackupCustomResource(ctx, s.T(), namespace, backup.GetClusterResourceName())
	// Check BackupJob cr deleted
	_, err = targetCluster.GetPDSBackupJob(ctx, namespace, backupJobName)
	expectedError := fmt.Sprintf("backupjobs.backups.pds.io %q not found", backupJobName)
	require.EqualError(s.T(), err, expectedError)
}

func (s *BackupJobTestSuite) DisconnectTC(replicas int32, deployment *appsv1.Deployment) {
	deployment.Spec.Replicas = &replicas
	_, err := targetCluster.UpdateDeployment(ctx, targetcluster.PDSChartNamespace, deployment)
	require.NoError(s.T(), err)
}
