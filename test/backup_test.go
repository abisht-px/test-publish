package test

import (
	"fmt"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/random"
)

func (s *PDSTestSuite) TestBackup_WithSchedule() {
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
	backupPolicyName1 := fmt.Sprintf("integration-test-%s", random.AlphaNumericString(random.NameSuffixLength))
	backupPolicyName2 := fmt.Sprintf("integration-test-%s", random.AlphaNumericString(random.NameSuffixLength))
	schedule1 := "*/1 * * * *"
	schedule2 := "*/2 * * * *"
	var retention int32 = 10
	backupPolicy1 := s.controlPlane.MustCreateBackupPolicy(s.ctx, s.T(), &backupPolicyName1, &schedule1, &retention)
	s.T().Cleanup(func() {
		_, _ = s.controlPlane.DeleteBackupPolicy(s.ctx, backupPolicy1.GetId())
	})
	backupPolicy2 := s.controlPlane.MustCreateBackupPolicy(s.ctx, s.T(), &backupPolicyName2, &schedule2, &retention)
	s.T().Cleanup(func() {
		_, _ = s.controlPlane.DeleteBackupPolicy(s.ctx, backupPolicy2.GetId())
	})

	// Deploy DS
	deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
	deployment.BackupPolicyname = *backupPolicy1.Name
	deployment.BackupTargetName = *backupTarget.Name
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, s.T(), &deployment)
	pdsDeployment, resp, err := s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespaceModel, resp, err := s.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(s.ctx, *pdsDeployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeploymentIfExists(s.ctx, s.T(), deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), deploymentID)
	})
	s.T().Cleanup(func() {
		// Cleanup scheduled backups and backupjobs
		backups := s.controlPlane.MustListBackupsByDeploymentID(s.ctx, s.T(), deploymentID)
		for _, backup := range backups {
			backupJobs, resp, err := s.controlPlane.ListBackupJobsInProject(s.ctx, s.controlPlane.TestPDSProjectID, controlplane.WithListBackupJobsInProjectBackupID(backup.GetId()))
			api.RequireNoError(s.T(), resp, err)
			for _, backupJob := range backupJobs {
				s.controlPlane.MustDeleteBackupJobByID(s.ctx, s.T(), backupJob.GetId())
			}
			s.controlPlane.MustDeleteBackup(s.ctx, s.T(), backup.GetId(), false)
		}
	})
	s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForStatefulSetReady(s.ctx, s.T(), deploymentID)

	// Wait for next 2 backups to trigger
	backups := s.controlPlane.MustListBackupsByDeploymentID(s.ctx, s.T(), deploymentID)
	s.Require().Equal(1, len(backups))
	scheduleBackup := backups[0]
	s.controlPlane.MustEnsureNBackupJobsSuccessFromSchedule(s.ctx, s.T(), s.controlPlane.TestPDSProjectID, scheduleBackup.GetId(), 2)
	cronjob, err := s.targetCluster.GetCronJob(s.ctx, namespace, *scheduleBackup.ClusterResourceName)
	s.Require().NoError(err)
	s.Require().Equal(cronjob.Spec.Schedule, schedule1)

	// When.
	// Update deployment with another backup schedule policy
	deployment.BackupPolicyname = backupPolicyName2
	s.controlPlane.MustUpdateDeployment(s.ctx, s.T(), deploymentID, &deployment)

	// Then.
	// Wait for new schedule to be created in TC
	backups = s.controlPlane.MustListBackupsByDeploymentID(s.ctx, s.T(), deploymentID)
	s.Require().Equal(2, len(backups))
	newScheduleBackup := backups[1] // get recent backup after backup policy update in deployment
	s.targetCluster.MustWaitForPDSBackupWithUpdatedSchedule(s.ctx, s.T(), namespace, *newScheduleBackup.ClusterResourceName, schedule2)
	// Wait for next 2 backup jobs to succeed
	s.controlPlane.MustEnsureNBackupJobsSuccessFromSchedule(s.ctx, s.T(), s.controlPlane.TestPDSProjectID, scheduleBackup.GetId(), 2)
}
