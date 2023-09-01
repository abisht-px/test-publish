package backup_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/backuptargets"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/random"
	"github.com/portworx/pds-integration-test/suites/framework"
)

func (s *BackupTestSuite) TestBackup_WithSchedule() {
	// Given
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Cassandra),
		NodeCount:       1,
	}
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
	// Setup backup policy
	backupPolicyName1 := fmt.Sprintf("integration-test-%s", random.AlphaNumericString(random.NameSuffixLength))
	backupPolicyName2 := fmt.Sprintf("integration-test-%s", random.AlphaNumericString(random.NameSuffixLength))
	schedule1 := "*/1 * * * *"
	schedule2 := "*/2 * * * *"
	var retention int32 = 10
	backupPolicy1 := controlPlane.MustCreateBackupPolicy(ctx, s.T(), &backupPolicyName1, &schedule1, &retention)
	s.T().Cleanup(func() {
		_, _ = controlPlane.DeleteBackupPolicy(ctx, backupPolicy1.GetId())
	})
	backupPolicy2 := controlPlane.MustCreateBackupPolicy(ctx, s.T(), &backupPolicyName2, &schedule2, &retention)
	s.T().Cleanup(func() {
		_, _ = controlPlane.DeleteBackupPolicy(ctx, backupPolicy2.GetId())
	})

	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
	deployment.BackupPolicyname = *backupPolicy1.Name
	deployment.BackupTargetName = *backupTarget.Name
	deploymentID := controlPlane.MustDeployDeploymentSpec(ctx, s.T(), &deployment)
	pdsDeployment, resp, err := controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeploymentIfExists(ctx, s.T(), deploymentID)
		controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)
	})
	s.T().Cleanup(func() {
		// Cleanup scheduled backups and backupjobs
		backups := controlPlane.MustListBackupsByDeploymentID(ctx, s.T(), deploymentID)
		for _, backup := range backups {
			backupJobs, resp, err := controlPlane.ListBackupJobsInProject(ctx, controlPlane.TestPDSProjectID, controlplane.WithListBackupJobsInProjectBackupID(backup.GetId()))
			api.RequireNoError(s.T(), resp, err)
			for _, backupJob := range backupJobs {
				controlPlane.MustDeleteBackupJobByID(ctx, s.T(), backupJob.GetId())
			}

			var apierr error
			s.Eventually(func() bool {
				resp, err := controlPlane.DeleteBackup(ctx, s.T(), backup.GetId(), false)
				if err != nil {
					apierr = api.ExtractErrorDetails(resp, err)
					return false
				}

				return true
			},
				framework.DefaultTimeout,
				framework.DefaultPollPeriod,
				apierr,
			)
		}
	})
	controlPlane.MustWaitForDeploymentHealthy(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForDeploymentInitialized(ctx, s.T(), deploymentID)
	crossCluster.MustWaitForStatefulSetReady(ctx, s.T(), deploymentID)

	// Wait for next 2 backups to trigger
	backups := controlPlane.MustListBackupsByDeploymentID(ctx, s.T(), deploymentID)
	s.Require().Equal(1, len(backups))
	scheduleBackup := backups[0]
	controlPlane.MustEnsureNBackupJobsSuccessFromSchedule(ctx, s.T(), controlPlane.TestPDSProjectID, scheduleBackup.GetId(), 2)
	cronjob, err := targetCluster.GetCronJob(ctx, namespace, *scheduleBackup.ClusterResourceName)
	s.Require().NoError(err)
	s.Require().Equal(cronjob.Spec.Schedule, schedule1)

	// When.
	// Update deployment with another backup schedule policy
	deployment.BackupPolicyname = backupPolicyName2
	controlPlane.MustUpdateDeployment(ctx, s.T(), deploymentID, &deployment)

	// Then.
	// Wait for new schedule to be created in TC
	backups = controlPlane.MustListBackupsByDeploymentID(ctx, s.T(), deploymentID)
	s.Require().Equal(2, len(backups))
	newScheduleBackup := backups[1] // get recent backup after backup policy update in deployment
	targetCluster.MustWaitForPDSBackupWithUpdatedSchedule(ctx, s.T(), namespace, *newScheduleBackup.ClusterResourceName, schedule2)
	// Wait for next 2 backup jobs to succeed
	controlPlane.MustEnsureNBackupJobsSuccessFromSchedule(ctx, s.T(), controlPlane.TestPDSProjectID, scheduleBackup.GetId(), 2)
}

func (s *BackupTestSuite) TestBackupData_AfterDeleteDeployment() {
	s.T().Skip("Disabled for DS-5978")

	// Given
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Cassandra),
		NodeCount:       1,
	}
	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
	deploymentID := controlPlane.MustDeployDeploymentSpec(ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		controlPlane.MustRemoveDeploymentIfExists(ctx, s.T(), deploymentID)
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
	s.T().Cleanup(func() {
		deleteBackupWithWorkaround(s.T(), backup, namespace)
	})

	backupJobCr, err := targetCluster.GetPDSBackupJob(ctx, namespace, fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName()))
	s.Require().NoError(err)
	backupJobId, err := getBackupJobID(backupJobCr)
	s.Require().NoError(err)
	backupJob := controlPlane.MustGetBackupJob(ctx, s.T(), backupJobId)
	s.Require().NotNil(backupJob.CloudSnapId)
	backupPathPrefix := getBackupPathPrefix(s.T(), *backupJob.CloudSnapId)
	provider := backuptargets.NewAwsS3StorageProvider(backupTargetConfig.Bucket, backupTargetConfig.Region, s3Creds.AccessKey, s3Creds.SecretKey)
	objsBeforeDeletion, err := provider.ListObjectsWithPrefix(backupPathPrefix)
	s.Require().NoError(err)
	s.Require().NotNil(objsBeforeDeletion)

	// When
	controlPlane.MustRemoveDeployment(ctx, s.T(), deploymentID)
	controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)

	// Then
	// Verify backup still exists in s3 storage bucket
	objsAfterDeletion, err := provider.ListObjectsWithPrefix(backupPathPrefix)
	s.Require().NoError(err)
	s.Require().NotNil(objsAfterDeletion)
	s.Require().Equal(len(objsBeforeDeletion.Contents), len(objsAfterDeletion.Contents))

	// Check VolumeSnapshot cr exists on TC
	volumeSnapshotName := fmt.Sprintf("%s-adhoc", backup.GetClusterResourceName())
	snapshotCr, err := targetCluster.GetVolumeSnapshot(ctx, namespace, volumeSnapshotName)
	s.Require().NoError(err)
	s.Require().NotNil(snapshotCr)
}

//nolint:unused
func getBackupPathPrefix(t *testing.T, cloudSnapID string) string {
	parts := strings.SplitN(cloudSnapID, "/", 2)
	require.Equalf(t, 2, len(parts), "invalid cloudsnap id")

	return parts[1]
}
