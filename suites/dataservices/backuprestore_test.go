package dataservices_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/suites/framework"
)

type BackupRestoreSuite struct {
	suite.Suite
	startTime time.Time

	controlPlane    *controlplane.ControlPlane
	targetCluster   *targetcluster.TargetCluster
	crossCluster    *crosscluster.CrossClusterHelper
	backupTargetCfg framework.BackupTargetConfig

	activeVersions framework.DSVersionMatrix
}

func (s *BackupRestoreSuite) SetupSuite() {
	s.startTime = time.Now()

	s.controlPlane, s.targetCluster, s.crossCluster = SetupSuite(
		s.T(),
		"ds-br",
		controlplane.WithAccountName(framework.PDSAccountName),
		controlplane.WithTenantName(framework.PDSTenantName),
		controlplane.WithProjectName(framework.PDSProjectName),
		controlplane.WithLoadImageVersions(),
		controlplane.WithCreateTemplatesAndStorageOptions(
			framework.NewRandomName("ds-br"),
		),
	)

	s.backupTargetCfg = framework.NewBackupTargetConfigFromFlags()

	activeVersions, err := framework.NewDSVersionMatrixFromFlags()
	require.NoError(s.T(), err, "Initialize dataservices version matrix")

	s.activeVersions = activeVersions
}

func (s *BackupRestoreSuite) TearDownSuite() {
	TearDownSuite(s.T(), s.controlPlane, s.targetCluster)
}

func (s *BackupRestoreSuite) TestDataService_BackupRestore() {
	if *skipBackups {
		s.T().Skip("Backup tests skipped.")
	}

	ctx := context.Background()

	backupEnabledServices := []string{
		dataservices.Cassandra,
		dataservices.Consul,
		dataservices.Couchbase,
		dataservices.ElasticSearch,
		dataservices.MongoDB,
		dataservices.MySQL,
		dataservices.Postgres,
		dataservices.Redis,
		// TODO(DS-5988): SQL Server backup jobs can't be registered to CP.
		// dataservices.SqlServer,
	}

	for _, dsName := range backupEnabledServices {
		versions := s.activeVersions.GetVersions(dsName)
		for _, version := range versions {
			nodeCounts := commonNodeCounts[dsName]

			// Test all node counts.
			for _, nodeCount := range nodeCounts {
				deployment := api.ShortDeploymentSpec{
					DataServiceName: dsName,
					ImageVersionTag: version,
					NodeCount:       nodeCount,
				}

				s.T().Run(fmt.Sprintf("backup-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
					if *skipBackupsMultinode && deployment.NodeCount > 1 {
						t.Skipf("Backup tests for the %d node %s data services is disabled.", deployment.NodeCount, deployment.DataServiceName)
					}

					t.Parallel()

					var backupCredentials *pds.ModelsBackupCredentials
					var backupTarget *pds.ModelsBackupTarget
					var backup *pds.ModelsBackup

					deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
					deploymentID := s.controlPlane.MustDeployDeploymentSpec(ctx, t, &deployment)
					namespace := s.controlPlane.MustGetNamespaceForDeployment(ctx, t, deploymentID)
					t.Cleanup(func() {
						s.controlPlane.MustRemoveDeploymentIfExists(ctx, t, deploymentID)
						s.crossCluster.MustDeleteDeploymentVolumes(ctx, t, deploymentID)
					})
					s.controlPlane.MustWaitForDeploymentHealthy(ctx, t, deploymentID)
					s.crossCluster.MustWaitForDeploymentInitialized(ctx, t, deploymentID)
					s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)

					seed := deploymentID
					s.crossCluster.MustRunWriteLoadTestJob(ctx, t, deploymentID, seed)

					// This is a temporary change and once DS-5768 is done this sleep can be removed
					if deployment.DataServiceName == dataservices.Couchbase {
						time.Sleep(200 * time.Second)
					}

					name := framework.NewRandomName("backup-creds")
					backupTargetConfig := s.backupTargetCfg
					s3Creds := backupTargetConfig.Credentials.S3
					backupCredentials = s.controlPlane.MustCreateS3BackupCredentials(ctx, t, s3Creds, name)
					t.Cleanup(func() { s.controlPlane.MustDeleteBackupCredentials(ctx, t, backupCredentials.GetId()) })

					backupTarget = s.controlPlane.MustCreateS3BackupTarget(ctx, t, backupCredentials.GetId(), backupTargetConfig.Bucket, backupTargetConfig.Region)
					t.Cleanup(func() { s.controlPlane.MustDeleteBackupTarget(ctx, t, backupTarget.GetId()) })
					s.controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, t, backupTarget.GetId())

					backup = s.controlPlane.MustCreateBackup(ctx, t, deploymentID, backupTarget.GetId())
					s.controlPlane.MustWaitForBackupCreated(ctx, t, backup.GetId())

					s.controlPlane.MustEnsureNBackupJobsSuccessFromSchedule(ctx, t, backup.GetProjectId(), backup.GetId(), 1)
					backupJobs := s.controlPlane.MustListBackupJobsInProject(
						ctx, t, backup.GetProjectId(),
						controlplane.WithListBackupJobsInProjectBackupID(backup.GetId()),
					)
					backupJobID := backupJobs[0].GetId()

					// Remove the original deployment to save resources.
					// TODO(DS-6032): can't delete a backup if deployment is not exist; enable after the fix.
					// s.controlPlane.MustRemoveDeployment(ctx, t, deploymentID)
					// s.controlPlane.MustWaitForDeploymentRemoved(ctx, t, deploymentID)

					restoreName := framework.NewRandomName("restore")
					restore := s.controlPlane.MustCreateRestore(ctx, t, backupJobID, restoreName, backup.GetNamespaceId(), backup.GetDeploymentTargetId())
					t.Cleanup(func() {
						s.controlPlane.MustDeleteBackup(ctx, t, backup.GetId(), false)
						s.controlPlane.MustWaitForBackupRemoved(ctx, t, backup.GetId())
						s.controlPlane.MustRemoveDeployment(ctx, t, restore.GetDeploymentId())
						s.controlPlane.MustWaitForDeploymentRemoved(ctx, t, restore.GetDeploymentId())
						s.crossCluster.MustDeleteDeploymentVolumes(ctx, t, restore.GetDeploymentId())
					})

					waitTimeout := dataservices.GetLongTimeoutFor(deployment.NodeCount)
					s.crossCluster.MustEnsureRestoreSuccessful(ctx, t, namespace, restore.GetClusterResourceName(), waitTimeout)
					s.crossCluster.MustWaitForStatefulSetPDSModeNormalReady(ctx, t, restore.GetDeploymentId())

					// This is a temporary change and once DS-5768 is done this sleep can be removed
					if deployment.DataServiceName == dataservices.Couchbase {
						time.Sleep(200 * time.Second)
					}

					// Run Read load test.
					s.crossCluster.MustRunReadLoadTestJob(ctx, t, restore.GetDeploymentId(), seed)

					// Run CRUD load test.
					s.crossCluster.MustRunLoadTestJob(ctx, t, restore.GetDeploymentId())
				})
			}
		}
	}
}
