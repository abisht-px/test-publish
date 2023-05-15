package test

import (
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

var (
	skipBackups          = flag.Bool("skip-backups", false, "Skip tests related to backups.")
	skipBackupsMultinode = flag.Bool("skip-backups-multinode", true, "Skip tests related to backups which are run on multi-node data services.")
)

func (s *PDSTestSuite) TestDataService_WriteData() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.14.0",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.5",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.7.1",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.8.0",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.1.1",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.10.9",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.31",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.5.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       1,
		},
		// SQL Server is not enabled yet.
		//{
		//	DataServiceName: dataservices.SqlServer,
		//	ImageVersionTag: "2019-CU18-rhel-8.5",
		//	NodeCount:       1,
		//},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("write-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("write-%s-n%d-", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.controlPlane.MustRemoveDeployment(s.ctx, t, deploymentID)
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
			})
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)

			s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_BackupRestore() {
	if *skipBackups {
		s.T().Skip("Backup tests skipped.")
	}
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "11.18",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "12.13",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "13.9",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.6",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.14.0",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.5",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.31",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.5.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       1,
		},
		// SQL Server is not enabled yet.
		//{
		//	DataServiceName: dataservices.SqlServer,
		//	ImageVersionTag: "2019-CU18-rhel-8.5",
		//	NodeCount:       1,
		//},
	}

	for _, d := range deployments {
		deployment := d

		if *skipBackupsMultinode && deployment.NodeCount > 1 {
			s.T().Skipf("Backup tests for the %d node %s data services is disabled.", deployment.NodeCount, deployment.DataServiceName)
		}

		s.T().Run(fmt.Sprintf("backup-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			restoreName := generateRandomName("restore")
			namespace := s.controlPlane.MustGetNamespaceForDeployment(s.ctx, t, deploymentID)
			t.Cleanup(func() {
				s.controlPlane.MustRemoveDeployment(s.ctx, t, deploymentID)
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)

				// TODO(DS-5494): that's a workaround, update it after a fix (ensure that restored DS is being deleted after deployment -
				//   	target operator reports deletion of data service CR to the control plane, it is based on the
				//      pds/deployment-id label in the CR; so if we delete the restored data service CR first, deployment in CP
				//      will be deleted as part of it and it breaks the overall cleanup process).
				err := s.targetCluster.DeletePDSDeployment(s.ctx, namespace, dataservices.ToPluralName(deployment.DataServiceName), restoreName)
				if !apierrors.IsNotFound(err) {
					require.NoError(t, err)
				}
			})
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)

			s.crossCluster.MustRunWriteLoadTestJob(s.ctx, t, deploymentID, "", deploymentID)

			name := generateRandomName("backup-creds")
			backupTargetConfig := s.config.backupTarget
			s3Creds := backupTargetConfig.credentials.S3
			backupCredentials := s.controlPlane.MustCreateS3BackupCredentials(s.ctx, t, s3Creds, name)
			t.Cleanup(func() { s.controlPlane.MustDeleteBackupCredentials(s.ctx, t, backupCredentials.GetId()) })

			backupTarget := s.controlPlane.MustCreateS3BackupTarget(s.ctx, t, backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
			s.controlPlane.MustEnsureBackupTargetCreatedInTC(s.ctx, t, backupTarget.GetId())
			t.Cleanup(func() { s.controlPlane.MustDeleteBackupTarget(s.ctx, t, backupTarget.GetId()) })

			backup := s.controlPlane.MustCreateBackup(s.ctx, t, deploymentID, backupTarget.GetId())
			s.crossCluster.MustEnsureBackupSuccessful(s.ctx, t, deploymentID, backup.GetClusterResourceName())
			t.Cleanup(func() { s.controlPlane.MustDeleteBackup(s.ctx, t, backup.GetId()) })

			if !isRestoreTestReadyFor(deployment.DataServiceName) {
				return
			}

			s.crossCluster.MustCreateRestore(s.ctx, t, deploymentID, backup.GetClusterResourceName(), restoreName)
			s.crossCluster.MustEnsureRestoreSuccessful(s.ctx, t, deploymentID, restoreName)
			t.Cleanup(func() {
				err := s.targetCluster.DeletePDSRestore(s.ctx, namespace, restoreName)
				require.NoError(t, err)
			})

			s.crossCluster.MustWaitForRestoredStatefulSetReady(s.ctx, t, deploymentID, restoreName)
			s.crossCluster.MustRunReadLoadTestJob(s.ctx, t, deploymentID, restoreName, deploymentID)

			s.crossCluster.MustRunCRUDLoadTestJob(s.ctx, t, deploymentID, restoreName, "")
		})
	}
}

func (s *PDSTestSuite) TestDataService_UpdateImage() {
	testCases := []struct {
		spec           api.ShortDeploymentSpec
		targetVersions []string
	}{
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.MongoDB,
				ImageVersionTag: "6.0.2",
				NodeCount:       1,
			},
			targetVersions: []string{"6.0.3"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "11.16",
				NodeCount:       1,
			},
			targetVersions: []string{"11.18"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "12.11",
				NodeCount:       1,
			},
			targetVersions: []string{"12.13"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "13.7",
				NodeCount:       1,
			},
			targetVersions: []string{"13.9"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "14.2",
				NodeCount:       1,
			},
			targetVersions: []string{"14.6"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "14.4",
				NodeCount:       1,
			},
			targetVersions: []string{"14.6"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "14.5",
				NodeCount:       1,
			},
			targetVersions: []string{"14.6"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "4.0.4",
				NodeCount:       1,
			},
			targetVersions: []string{"4.0.6"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "4.0.5",
				NodeCount:       1,
			},
			targetVersions: []string{"4.0.6"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Redis,
				ImageVersionTag: "7.0.0",
				NodeCount:       1,
			},
			targetVersions: []string{"7.0.5"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Redis,
				ImageVersionTag: "7.0.2",
				NodeCount:       1,
			},
			targetVersions: []string{"7.0.5"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Redis,
				ImageVersionTag: "7.0.4",
				NodeCount:       1,
			},
			targetVersions: []string{"7.0.5"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.2.0",
				NodeCount:       1,
			},
			targetVersions: []string{"3.2.3"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.2.1",
				NodeCount:       1,
			},
			targetVersions: []string{"3.2.3"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.RabbitMQ,
				ImageVersionTag: "3.9.21",
				NodeCount:       1,
			},
			targetVersions: []string{"3.9.22"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.RabbitMQ,
				ImageVersionTag: "3.10.6",
				NodeCount:       1,
			},
			targetVersions: []string{"3.10.9"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.RabbitMQ,
				ImageVersionTag: "3.10.7",
				NodeCount:       1,
			},
			targetVersions: []string{"3.10.9"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.MySQL,
				ImageVersionTag: "8.0.30",
				NodeCount:       1,
			},
			targetVersions: []string{"8.0.31"},
		},
	}

	for _, testCase := range testCases {
		for _, tvt := range testCase.targetVersions {
			tt := testCase
			targetVersionTag := tvt
			s.T().Run(fmt.Sprintf("update-%s-%s-to-%s", tt.spec.DataServiceName, tt.spec.ImageVersionString(), targetVersionTag), func(t *testing.T) {
				t.Parallel()

				tt.spec.NamePrefix = fmt.Sprintf("update-%s-", tt.spec.ImageVersionString())
				deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &tt.spec)
				t.Cleanup(func() {
					s.controlPlane.MustRemoveDeployment(s.ctx, t, deploymentID)
					s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
				})

				// Create.
				s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
				s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, t, deploymentID)
				s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)
				s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)

				// Update.
				newSpec := tt.spec
				newSpec.ImageVersionTag = targetVersionTag
				oldUpdateRevision := s.crossCluster.MustGetStatefulSetUpdateRevision(s.ctx, t, deploymentID)
				s.controlPlane.MustUpdateDeployment(s.ctx, t, deploymentID, &newSpec)
				s.crossCluster.MustWaitForStatefulSetChanged(s.ctx, t, deploymentID, oldUpdateRevision)
				s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
				s.crossCluster.MustWaitForStatefulSetImage(s.ctx, t, deploymentID, targetVersionTag)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)
				s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)
			})
		}
	}
}

func (s *PDSTestSuite) TestDataService_ScaleUp() {
	testCases := []struct {
		spec    api.ShortDeploymentSpec
		scaleTo int
	}{
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "11.18",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "12.13",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "13.9",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "14.6",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "4.0.6",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Consul,
				ImageVersionTag: "1.14.0",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Redis,
				ImageVersionTag: "7.0.5",
				NodeCount:       6,
			},
			scaleTo: 8,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.1.1",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.2.3",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.RabbitMQ,
				ImageVersionTag: "3.10.9",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.MySQL,
				ImageVersionTag: "8.0.31",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.MongoDB,
				ImageVersionTag: "6.0.3",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.ElasticSearch,
				ImageVersionTag: "8.5.2",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Couchbase,
				ImageVersionTag: "7.1.1",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
	}

	for _, testCase := range testCases {
		tt := testCase
		s.T().Run(fmt.Sprintf("scale-%s-%s-nodes-%v-to-%v", tt.spec.DataServiceName, tt.spec.ImageVersionString(), tt.spec.NodeCount, tt.scaleTo), func(t *testing.T) {
			t.Parallel()

			tt.spec.NamePrefix = fmt.Sprintf("scale-%s-", tt.spec.ImageVersionString())
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &tt.spec)
			t.Cleanup(func() {
				s.controlPlane.MustRemoveDeployment(s.ctx, t, deploymentID)
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
			})

			// Create.
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)
			s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)

			// Update.
			updateSpec := tt.spec
			updateSpec.NodeCount = tt.scaleTo
			oldUpdateRevision := s.crossCluster.MustGetStatefulSetUpdateRevision(s.ctx, t, deploymentID)
			s.controlPlane.MustUpdateDeployment(s.ctx, t, deploymentID, &updateSpec)
			s.crossCluster.MustWaitForStatefulSetChanged(s.ctx, t, deploymentID, oldUpdateRevision)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)
			s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_ScaleResources() {
	testCases := []struct {
		spec                    api.ShortDeploymentSpec
		scaleToResourceTemplate string
	}{
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "14.6",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.Postgres].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "4.0.6",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.Cassandra].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Consul,
				ImageVersionTag: "1.14.0",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.Consul].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.2.3",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.Kafka].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.RabbitMQ,
				ImageVersionTag: "3.10.9",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.RabbitMQ].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.MySQL,
				ImageVersionTag: "8.0.31",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.MySQL].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.MongoDB,
				ImageVersionTag: "6.0.2",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.MongoDB].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.ElasticSearch,
				ImageVersionTag: "8.5.2",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.ElasticSearch].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Couchbase,
				ImageVersionTag: "7.1.1",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.Couchbase].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Redis,
				ImageVersionTag: "7.0.5",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.Redis].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.ZooKeeper,
				ImageVersionTag: "3.8.0",
				NodeCount:       3,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.ZooKeeper].ResourceTemplates[1].Name,
		},
		// SQL Server is not enabled yet.
		//{
		//	spec: api.ShortDeploymentSpec{
		//		DataServiceName: dataservices.SqlServer,
		//		ImageVersionTag: "2019-CU18-rhel-8.5",
		//		NodeCount:       1,
		//	},
		//	scaleToResourceTemplate: s.controlPlane.TestPDSTemplates[dataservices.SqlServer].ResourceTemplates[1].Name,
		//},
	}

	for _, testCase := range testCases {
		tt := testCase
		s.T().Run(fmt.Sprintf("scale-%s-%s-resources", tt.spec.DataServiceName, tt.spec.ImageVersionString()), func(t *testing.T) {
			t.Parallel()

			tt.spec.NamePrefix = fmt.Sprintf("scale-%s-", tt.spec.ImageVersionString())
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &tt.spec)
			t.Cleanup(func() {
				s.controlPlane.MustRemoveDeployment(s.ctx, t, deploymentID)
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
			})

			// Create.
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)
			s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)

			// Update.
			updateSpec := tt.spec
			updateSpec.ResourceSettingsTemplateName = tt.scaleToResourceTemplate
			oldUpdateRevision := s.crossCluster.MustGetStatefulSetUpdateRevision(s.ctx, t, deploymentID)
			s.controlPlane.MustUpdateDeployment(s.ctx, t, deploymentID, &updateSpec)
			s.crossCluster.MustWaitForStatefulSetChanged(s.ctx, t, deploymentID, oldUpdateRevision)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)
			s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_Recovery_FromDeletion() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.6",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.14.0",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.6",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.5",
			NodeCount:       6,
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.8.0",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.10.9",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.31",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.5.2",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       3,
		},
		// SQL Server is not enabled yet.
		//{
		//	DataServiceName: dataservices.SqlServer,
		//	ImageVersionTag: "2019-CU18-rhel-8.5",
		//	NodeCount:       1,
		//},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("recover-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("recover-%s-n%d-", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.controlPlane.MustRemoveDeployment(s.ctx, t, deploymentID)
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
			})
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)
			s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)
			//Delete pods and load test
			s.targetCluster.MustDeleteDeploymentPods(s.ctx, t, s.config.pdsNamespaceName, deploymentID)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)
			s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_Metrics() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.6",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       3,
		},
		// TODO: https://portworx.atlassian.net/browse/DS-4878
		// {
		// 	DataServiceName: dataservices.Consul,
		// 	ImageVersionTag: "1.14.0",
		// 	NodeCount:       3,
		// },
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.31",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.5.2",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.10.9",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.5",
			NodeCount:       6,
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.8.0",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.6",
			NodeCount:       1,
		},
		// SQL Server is not enabled yet.
		//{
		//	DataServiceName: dataservices.SqlServer,
		//	ImageVersionTag: "2019-CU18-rhel-8.5",
		//	NodeCount:       1,
		//},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("metrics-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("metrics-%s-n%d-", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.controlPlane.MustRemoveDeployment(s.ctx, t, deploymentID)
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
			})
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)
			s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)

			// Try to get DS metrics from prometheus.
			s.controlPlane.MustWaitForMetricsReported(s.ctx, t, deploymentID)
		})
	}
}

func isRestoreTestReadyFor(dataServiceName string) bool {
	return dataServiceName == dataservices.Postgres
}
