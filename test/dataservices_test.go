package test

import (
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/strings/slices"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/kubernetes/psa"
	"github.com/portworx/pds-integration-test/internal/random"
)

var (
	latestCompatibleOnly = flag.Bool("latest-compatible-only", true, "Test only update to the latest compatible version.")
	skipBackups          = flag.Bool("skip-backups", false, "Skip tests related to backups.")
	skipBackupsMultinode = flag.Bool("skip-backups-multinode", true, "Skip tests related to backups which are run on multi-node data services.")
)

func (s *PDSTestSuite) TestDataService_DeploymentWithPSA() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.1.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.10",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.11.15",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.0.29",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.15.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.14.7",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.8.0",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.4.1",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.3.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.1.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.33",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "15.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.8",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "13.11",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "12.15",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "11.20",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.11.16",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.10.22",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.9",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.SqlServer,
			ImageVersionTag: "2019-CU20",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.7.1",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.8.1",
			NodeCount:       3,
		},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("deploy-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			// Create namespace with PSA policy set
			psaPolicy := getSupportedPSAPolicy(deployment.DataServiceName)
			namespaceName := "it-" + psaPolicy + "-" + random.AlphaNumericString(4)
			namespace := psa.NewNamespace(namespaceName, psaPolicy, true)
			_, err := s.targetCluster.CreateNamespace(s.ctx, namespace)
			t.Cleanup(func() {
				_ = s.targetCluster.DeleteNamespace(s.ctx, namespaceName)
			})
			s.Require().NoError(err)
			modelsNamespace := s.controlPlane.MustWaitForNamespaceStatus(s.ctx, t, namespaceName, "available")
			namespaceID := modelsNamespace.GetId()

			deployment.NamePrefix = fmt.Sprintf("deploy-%s-n%d-", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpecIntoNamespace(s.ctx, t, &deployment, namespaceID)
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
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.1.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.1.2",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.10",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.10",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.11.15",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.11.15",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.0.29",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.0.29",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.15.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.15.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.14.7",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.14.7",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.8.0",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.8.0",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.6",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.33",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.33",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "15.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "15.3",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.8",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.8",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "13.11",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "13.11",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "12.15",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "12.15",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "11.20",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "11.20",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.9",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.9",
			NodeCount:       6,
		},
		{
			DataServiceName: dataservices.SqlServer,
			ImageVersionTag: "2019-CU20",
			NodeCount:       1,
		},
	}

	for _, d := range deployments {
		deployment := d

		s.T().Run(fmt.Sprintf("backup-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			if *skipBackupsMultinode && deployment.NodeCount > 1 {
				t.Skipf("Backup tests for the %d node %s data services is disabled.", deployment.NodeCount, deployment.DataServiceName)
			}

			t.Parallel()

			var backupCredentials *pds.ModelsBackupCredentials
			var backupTarget *pds.ModelsBackupTarget
			var backup *pds.ModelsBackup
			var restoreCreated = false

			deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			restoreName := generateRandomName("restore")
			namespace := s.controlPlane.MustGetNamespaceForDeployment(s.ctx, t, deploymentID)
			t.Cleanup(func() {
				if backup != nil {
					deleteBackupWithWorkaround(s, t, backup, namespace)
				}
				if backupTarget != nil {
					s.controlPlane.MustDeleteBackupTarget(s.ctx, t, backupTarget.GetId())
				}
				if backupCredentials != nil {
					s.controlPlane.MustDeleteBackupCredentials(s.ctx, t, backupCredentials.GetId())
				}

				s.controlPlane.MustRemoveDeploymentIfExists(s.ctx, t, deploymentID)
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)

				if restoreCreated {
					err := s.targetCluster.DeletePDSRestore(s.ctx, namespace, restoreName)
					require.NoError(t, err)
				}

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

			seed := deploymentID
			s.crossCluster.MustRunWriteLoadTestJob(s.ctx, t, deploymentID, seed)

			// This is a temporary change and once DS-5768 is done this sleep can be removed
			if deployment.DataServiceName == dataservices.Couchbase {
				time.Sleep(200 * time.Second)
			}

			name := generateRandomName("backup-creds")
			backupTargetConfig := s.config.backupTarget
			s3Creds := backupTargetConfig.credentials.S3
			backupCredentials = s.controlPlane.MustCreateS3BackupCredentials(s.ctx, t, s3Creds, name)

			backupTarget = s.controlPlane.MustCreateS3BackupTarget(s.ctx, t, backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
			s.controlPlane.MustEnsureBackupTargetCreatedInTC(s.ctx, t, backupTarget.GetId())

			// Create backup (with retry if needed).
			for i := 1; i <= 3; i++ {
				backup = s.controlPlane.MustCreateBackup(s.ctx, t, deploymentID, backupTarget.GetId())
				needsRetry := s.crossCluster.MustEnsureBackupSuccessful(s.ctx, t, deploymentID, backup.GetClusterResourceName())
				if needsRetry {
					// Delete failed backup before retry.
					backupToDelete := backup
					backup = nil
					deleteBackupWithWorkaround(s, t, backupToDelete, namespace)
					// Wait a bit then repeat.
					time.Sleep(15 * time.Second)
				} else {
					break
				}
			}

			if !isRestoreTestReadyFor(deployment.DataServiceName) {
				return
			}

			// Remove the original deployment to save resources.
			s.controlPlane.MustRemoveDeployment(s.ctx, t, deploymentID)
			s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)

			s.crossCluster.MustCreateRestore(s.ctx, t, namespace, backup.GetClusterResourceName(), restoreName)
			restoreCreated = true
			waitTimeout := dataservices.GetLongTimeoutFor(deployment.NodeCount)
			s.crossCluster.MustEnsureRestoreSuccessful(s.ctx, t, namespace, restoreName, waitTimeout)

			s.crossCluster.MustWaitForStatefulSetInPDSModeNormal(s.ctx, t, namespace, restoreName)
			s.crossCluster.MustWaitForRestoredStatefulSetReady(s.ctx, t, namespace, restoreName, deployment.NodeCount)

			// Temporary fix for MySQL - can remove after DS-4984 is completed.
			if deployment.DataServiceName == dataservices.MySQL {
				time.Sleep(200 * time.Second)
			}

			// Run Read load test.
			s.crossCluster.MustRunGenericLoadTestJob(s.ctx, t, deployment.DataServiceName, namespace, restoreName, crosscluster.LoadTestRead, seed, crosscluster.PDSUser, deployment.NodeCount, nil)

			// Run CRUD load test.
			s.crossCluster.MustRunGenericLoadTestJob(s.ctx, t, deployment.DataServiceName, namespace, restoreName, crosscluster.LoadTestCRUD, "", crosscluster.PDSUser, deployment.NodeCount, nil)
		})
	}
}

func (s *PDSTestSuite) TestDataService_UpdateImage() {

	dataServices := []string{
		dataservices.Cassandra,
		dataservices.Couchbase,
		dataservices.Kafka,
		dataservices.MongoDB,
		dataservices.MySQL,
		dataservices.Postgres,
		dataservices.RabbitMQ,
		dataservices.Redis,
		dataservices.SqlServer,
		dataservices.ZooKeeper,
		dataservices.ElasticSearch,
		dataservices.Consul,
	}

	compatibleVersions := s.controlPlane.MustGetCompatibleVersions(s.ctx, s.T())

	for _, cv := range compatibleVersions {
		dataServiceName := *cv.DataServiceName
		// Filter for selected data services only.
		if !slices.Contains(dataServices, dataServiceName) {
			continue
		}

		// TODO Use Nick's minNodeCount map for this, once available
		nodeCount := int32(1)
		if dataServiceName == dataservices.ZooKeeper {
			nodeCount = 3
		}

		targets := cv.Compatible
		if *latestCompatibleOnly {
			targets = cv.LatestCompatible
		}
		for _, target := range targets {
			spec := api.ShortDeploymentSpec{
				DataServiceName: dataServiceName,
				ImageVersionTag: *cv.VersionName,
				NodeCount:       nodeCount,
			}
			targetVersionTag := *target.Name
			s.T().Run(fmt.Sprintf("update-%s-%s-to-%s", spec.DataServiceName, spec.ImageVersionString(), targetVersionTag), func(t *testing.T) {
				t.Parallel()

				spec.NamePrefix = fmt.Sprintf("update-%s-", spec.ImageVersionString())
				deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &spec)
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
				newSpec := spec
				newSpec.ImageVersionTag = targetVersionTag
				oldUpdateRevision := s.crossCluster.MustGetStatefulSetUpdateRevision(s.ctx, t, deploymentID)
				s.controlPlane.MustUpdateDeployment(s.ctx, t, deploymentID, &newSpec)
				s.crossCluster.MustWaitForStatefulSetChanged(s.ctx, t, deploymentID, oldUpdateRevision)
				s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
				s.crossCluster.MustWaitForStatefulSetImage(s.ctx, t, deploymentID, targetVersionTag)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)

				// Temporary fix for MySQL - can remove after DS-4984 is completed.
				if spec.DataServiceName == dataservices.MySQL {
					time.Sleep(200 * time.Second)
				}
				s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)
			})
		}
	}
}

func (s *PDSTestSuite) TestDataService_ScaleUp() {
	testCases := []struct {
		spec    api.ShortDeploymentSpec
		scaleTo int32
	}{
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "4.1.2",
				NodeCount:       2,
			},
			scaleTo: 3,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "4.0.10",
				NodeCount:       2,
			},
			scaleTo: 3,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "3.11.15",
				NodeCount:       2,
			},
			scaleTo: 3,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "3.0.29",
				NodeCount:       2,
			},
			scaleTo: 3,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Consul,
				ImageVersionTag: "1.15.3",
				NodeCount:       1,
			},
			scaleTo: 3,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Consul,
				ImageVersionTag: "1.14.7",
				NodeCount:       1,
			},
			scaleTo: 3,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Couchbase,
				ImageVersionTag: "7.1.1",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.ElasticSearch,
				ImageVersionTag: "8.8.0",
				NodeCount:       1,
			},
			scaleTo: 3,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.4.1",
				NodeCount:       3,
			},
			scaleTo: 5,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.3.2",
				NodeCount:       3,
			},
			scaleTo: 5,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.2.3",
				NodeCount:       3,
			},
			scaleTo: 5,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.1.2",
				NodeCount:       3,
			},
			scaleTo: 5,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.MongoDB,
				ImageVersionTag: "6.0.6",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.MySQL,
				ImageVersionTag: "8.0.33",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "15.3",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "14.8",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "13.11",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "12.15",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "11.20",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.RabbitMQ,
				ImageVersionTag: "3.11.16",
				NodeCount:       1,
			},
			scaleTo: 3,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.RabbitMQ,
				ImageVersionTag: "3.10.22",
				NodeCount:       1,
			},
			scaleTo: 3,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Redis,
				ImageVersionTag: "7.0.9",
				NodeCount:       6,
			},
			scaleTo: 8,
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

			// Temporary fix for MySQL - can remove after DS-4984 is completed.
			if tt.spec.DataServiceName == dataservices.MySQL {
				time.Sleep(200 * time.Second)
			}
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
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "4.1.2",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "4.0.10",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "3.11.15",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Cassandra,
				ImageVersionTag: "3.0.29",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Consul,
				ImageVersionTag: "1.15.3",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Consul,
				ImageVersionTag: "1.14.7",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Couchbase,
				ImageVersionTag: "7.1.1",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.ElasticSearch,
				ImageVersionTag: "8.8.0",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.4.1",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.3.2",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.2.3",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Kafka,
				ImageVersionTag: "3.1.2",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.MongoDB,
				ImageVersionTag: "6.0.6",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.MySQL,
				ImageVersionTag: "8.0.33",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "15.3",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "14.8",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "13.11",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "12.15",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Postgres,
				ImageVersionTag: "11.20",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.RabbitMQ,
				ImageVersionTag: "3.11.16",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.RabbitMQ,
				ImageVersionTag: "3.10.22",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.Redis,
				ImageVersionTag: "7.0.9",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.SqlServer,
				ImageVersionTag: "2019-CU20",
				NodeCount:       1,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.ZooKeeper,
				ImageVersionTag: "3.8.1",
				NodeCount:       3,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dataservices.ZooKeeper,
				ImageVersionTag: "3.7.1",
				NodeCount:       3,
			},
			scaleToResourceTemplate: dataservices.TemplateNameMed,
		},
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

			// Temporary fix for MySQL - can remove after DS-4984 is completed.
			if tt.spec.DataServiceName == dataservices.MySQL {
				time.Sleep(200 * time.Second)
			}

			s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_Recovery_FromDeletion() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.1.2",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.10",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.11.15",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.0.29",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.15.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.14.7",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.8.0",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.4.1",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.3.2",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.1.2",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.6",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.33",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "15.3",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.8",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "13.11",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "12.15",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "11.20",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.11.16",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.10.22",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.9",
			NodeCount:       6,
		},
		{
			DataServiceName: dataservices.SqlServer,
			ImageVersionTag: "2019-CU20",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.7.1",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.8.1",
			NodeCount:       3,
		},
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

			// Temporary fix for MySQL - can remove after DS-4984 is completed.
			if deployment.DataServiceName == dataservices.MySQL {
				time.Sleep(200 * time.Second)
			}

			s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_ImpossibleResourceAllocation_Fails() {
	deployment := api.ShortDeploymentSpec{
		DataServiceName:              dataservices.Cassandra,
		NamePrefix:                   "impossible-resources-test",
		ImageVersionTag:              "4.1.2",
		NodeCount:                    3,
		ResourceSettingsTemplateName: dataservices.TemplateNameEnormous,
	}
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), deploymentID)
	})

	s.controlPlane.MustWaitForDeploymentEventCondition(s.ctx, s.T(),
		deploymentID,
		func(event pds.ModelsDeploymentTargetDeploymentEvent) bool {
			if event.Reason == nil {
				return false
			}
			reason := *event.Reason
			message := *event.Message
			insufficientResources := strings.Contains(message, "Insufficient cpu") || strings.Contains(message, "Insufficient memory")
			return reason == "FailedScheduling" && insufficientResources
		},
		"failed pod scheduling")
}

func (s *PDSTestSuite) TestDataService_Metrics() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.1.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.10",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.11.15",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.0.29",
			NodeCount:       1,
		},
		// TODO: https://portworx.atlassian.net/browse/DS-4878
		//{
		//	DataServiceName: dataservices.Consul,
		//	ImageVersionTag: "1.15.3",
		//	NodeCount:       1,
		//},
		//{
		//	DataServiceName: dataservices.Consul,
		//	ImageVersionTag: "1.14.7",
		//	NodeCount:       1,
		//},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.8.0",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.4.1",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.3.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.1.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.6",
			NodeCount:       2,
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.33",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "15.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.8",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "13.11",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "12.15",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "11.20",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.11.16",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.10.22",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.9",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.SqlServer,
			ImageVersionTag: "2019-CU20",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.7.1",
			NodeCount:       3,
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.8.1",
			NodeCount:       3,
		},
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

func (s *PDSTestSuite) TestDataService_DeletePDSUser() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.1.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.10",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.11.15",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "3.0.29",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.33",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "15.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.8",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "13.11",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "12.15",
			NodeCount:       1,
		},
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "11.20",
			NodeCount:       1,
		},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("userdel-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("userdel-%s-n%d-", deployment.ImageVersionString(), deployment.NodeCount)
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

			// Delete 'pds' user.
			s.crossCluster.MustRunDeleteUserJob(s.ctx, t, deploymentID, crosscluster.PDSUser)
			// Run CRUD tests with 'pds' to check that the data service fails (user does not exist).
			s.crossCluster.MustRunCRUDLoadTestJobAndFail(s.ctx, t, deploymentID, crosscluster.PDSUser)
			// Wait 30s before the check whether the pod was not killed due to readiness/liveness failure.
			time.Sleep(30 * time.Second)
			// Run CRUD tests with 'pds_replace_user' to check that the data service still works.
			s.crossCluster.MustRunCRUDLoadTestJob(s.ctx, t, deploymentID, crosscluster.PDSReplaceUser)
		})
	}
}

func isRestoreTestReadyFor(dataServiceName string) bool {
	switch dataServiceName {
	case dataservices.Cassandra,
		dataservices.Consul,
		dataservices.ElasticSearch,
		dataservices.MongoDB,
		dataservices.Postgres,
		dataservices.MySQL,
		dataservices.Redis,
		dataservices.Couchbase,
		dataservices.SqlServer:
		return true
	}
	return false
}

func getSupportedPSAPolicy(dataServiceName string) string {
	// https://pds.docs.portworx.com/concepts/pod-security-admission/#supported-security-levels-for-pds-resources
	switch dataServiceName {
	case dataservices.Cassandra:
		return psa.PSAPolicyPrivileged
	default:
		return psa.PSAPolicyRestricted
	}
}

func deleteBackupWithWorkaround(s *PDSTestSuite, t *testing.T, backup *pds.ModelsBackup, namespace string) {
	// TODO(DS-5732): Once bug https://portworx.atlassian.net/browse/DS-5732 is fixed then call MustDeleteBackup
	// 		with localOnly=false and remove the additional call of DeletePDSBackup.
	s.controlPlane.MustDeleteBackup(s.ctx, t, backup.GetId(), true)
	err := s.targetCluster.DeletePDSBackup(s.ctx, namespace, backup.GetClusterResourceName())
	require.NoError(t, err)
}
