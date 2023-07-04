package test

import (
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/kubernetes/psa"
	"github.com/portworx/pds-integration-test/internal/random"
)

const pdsSystemUsersCapabilityName = "pds_system_users"

var (
	latestCompatibleOnly = flag.Bool("latest-compatible-only", true, "Test only update to the latest compatible version.")
	skipBackups          = flag.Bool("skip-backups", false, "Skip tests related to backups.")
	skipBackupsMultinode = flag.Bool("skip-backups-multinode", true, "Skip tests related to backups which are run on multi-node data services.")

	// Modify this map to control which data services and versions are tested - all tests use this map.
	activeVersions = map[string][]string{
		dataservices.Cassandra:     {"4.1.2", "4.0.10", "3.11.15", "3.0.29"},
		dataservices.Consul:        {"1.15.3", "1.14.7"},
		dataservices.Couchbase:     {"7.1.1"},
		dataservices.ElasticSearch: {"8.8.0"},
		dataservices.Kafka:         {"3.4.1", "3.3.2", "3.2.3", "3.1.2"},
		dataservices.MongoDB:       {"6.0.6"},
		dataservices.MySQL:         {"8.0.33"},
		dataservices.Postgres:      {"15.3", "14.8", "13.11", "12.15", "11.20"},
		dataservices.RabbitMQ:      {"3.11.16", "3.10.22"},
		dataservices.Redis:         {"7.0.9"},
		dataservices.SqlServer:     {"2019-CU20"},
		dataservices.ZooKeeper:     {"3.8.1", "3.7.1"},
	}

	// Common node counts per data service.
	// Tests can decide whether to use the low count, the high count, or both.
	// Assumption: counts are sorted in ascending order.
	commonNodeCounts = map[string][]int32{
		dataservices.Cassandra:     {1, 3},
		dataservices.Consul:        {1, 3},
		dataservices.Couchbase:     {1, 2},
		dataservices.ElasticSearch: {1, 3},
		dataservices.Kafka:         {1, 3},
		dataservices.MongoDB:       {1, 2},
		dataservices.MySQL:         {1, 2},
		dataservices.Postgres:      {1, 2},
		dataservices.RabbitMQ:      {1, 3},
		dataservices.Redis:         {1, 6},
		dataservices.SqlServer:     {1},
		dataservices.ZooKeeper:     {3},
	}
)

func (s *PDSTestSuite) TestDataService_DeploymentWithPSA() {
	for dsName, versions := range activeVersions {
		for _, version := range versions {
			nodeCounts := commonNodeCounts[dsName]
			if len(nodeCounts) == 0 {
				continue
			}

			deployment := api.ShortDeploymentSpec{
				DataServiceName: dsName,
				ImageVersionTag: version,

				// Only test lowest node count.
				NodeCount: nodeCounts[0],
			}

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
}

func (s *PDSTestSuite) TestDataService_BackupRestore() {
	if *skipBackups {
		s.T().Skip("Backup tests skipped.")
	}

	backupEnabledServices := []string{
		dataservices.Cassandra,
		dataservices.Consul,
		dataservices.Couchbase,
		dataservices.ElasticSearch,
		dataservices.MongoDB,
		dataservices.MySQL,
		dataservices.Postgres,
		dataservices.Redis,
		dataservices.SqlServer,
	}

	for _, dsName := range backupEnabledServices {
		versions := activeVersions[dsName]
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
	}
}

func (s *PDSTestSuite) TestDataService_UpdateImage() {

	compatibleVersions := s.controlPlane.MustGetCompatibleVersions(s.ctx, s.T())
	for _, cv := range compatibleVersions {
		dataServiceName := *cv.DataServiceName

		// Filter for selected data services only.
		_, ok := activeVersions[dataServiceName]
		if !ok {
			continue
		}

		nodeCounts := commonNodeCounts[dataServiceName]
		if len(nodeCounts) == 0 {
			continue
		}

		fromSpec := api.ShortDeploymentSpec{
			DataServiceName: dataServiceName,
			ImageVersionTag: *cv.VersionName,
			// Only test lowest node count.
			NodeCount: nodeCounts[0],
		}
		s.controlPlane.SetDefaultImageVersionBuild(&fromSpec, false)

		targets := cv.Compatible
		if *latestCompatibleOnly {
			targets = cv.LatestCompatible
		}
		for _, target := range targets {
			fromSpec.NamePrefix = fmt.Sprintf("update-%s-", fromSpec.ImageVersionTag)
			toSpec := fromSpec
			toSpec.ImageVersionTag = *target.Name
			s.controlPlane.SetDefaultImageVersionBuild(&toSpec, true)

			testName := fmt.Sprintf("update-%s-%s-to-%s", dataServiceName, fromSpec.ImageVersionString(), toSpec.ImageVersionString())
			s.T().Run(testName, func(t *testing.T) {
				t.Parallel()
				s.updateTestImpl(t, fromSpec, toSpec)
			})
		}
	}
}

func (s *PDSTestSuite) TestDataService_PDSSystemUsersV1Migration() {
	dataServicesByName := s.controlPlane.MustGetDataServicesByName(s.ctx, s.T())
	for dsName, versions := range activeVersions {
		if dsName == dataservices.SqlServer {
			// No need to test migration for SQL Server, as no deployments exist on Prod and Staging
			// using older images without the PDS System Users V1 feature.
			continue
		}
		dataService, ok := dataServicesByName[dsName]
		if !ok {
			assert.Fail(s.T(), "Data service with name '%s' not found", dsName)
		}
		dsImages := s.controlPlane.MustGetAllImagesForDataService(s.ctx, s.T(), dataService.GetId())
	versionLoop:
		for _, version := range versions {
			nodeCounts := commonNodeCounts[dsName]
			if len(nodeCounts) == 0 {
				continue
			}

			toSpec := api.ShortDeploymentSpec{
				DataServiceName: dsName,
				ImageVersionTag: version,
				// Only test lowest node count.
				NodeCount: nodeCounts[0],
			}
			s.controlPlane.SetDefaultImageVersionBuild(&toSpec, false)

			// Find the build to migrate from.
			versionNamePrefix := getPatchVersionNamePrefix(dsName, toSpec.ImageVersionTag)
			filteredImages := filterImagesByVersionNamePrefix(dsImages, versionNamePrefix)
			var fromImage *pds.ModelsImage
			toImageFound := false
			for _, image := range filteredImages {
				// First find image for toSpec.
				if !toImageFound {
					if *image.Tag == toSpec.ImageVersionTag && *image.Build == toSpec.ImageVersionBuild {
						toImageFound = true
						if !hasPDSSystemUsersCapability(image) {
							s.T().Logf("Image %s %s does not have PDSSystemUsers capability defined.", dsName, toSpec.ImageVersionString())
							continue versionLoop
						}
					}
					continue
				}
				// Next find the latest image which does not have "pds_system_users" capability defined.
				if !hasPDSSystemUsersCapability(image) {
					fromImage = &image
					break
				}
			}
			if fromImage == nil {
				s.T().Logf("No previous image found without PDSSystemUsers capability %s %s", dsName, toSpec.ImageVersionString())
				continue
			}

			toSpec.NamePrefix = fmt.Sprintf("migrate-%s-", toSpec.ImageVersionTag)
			fromSpec := toSpec
			fromSpec.ImageVersionTag = *fromImage.Tag
			fromSpec.ImageVersionBuild = *fromImage.Build

			testName := fmt.Sprintf("migrate-%s-%s-to-%s", dsName, fromSpec.ImageVersionString(), toSpec.ImageVersionString())
			s.T().Run(testName, func(t *testing.T) {
				t.Parallel()
				s.updateTestImpl(t, fromSpec, toSpec)
			})
		}
	}
}

func (s *PDSTestSuite) updateTestImpl(t *testing.T, fromSpec, toSpec api.ShortDeploymentSpec) {
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &fromSpec)
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
	oldUpdateRevision := s.crossCluster.MustGetStatefulSetUpdateRevision(s.ctx, t, deploymentID)
	s.controlPlane.MustUpdateDeployment(s.ctx, t, deploymentID, &toSpec)
	s.crossCluster.MustWaitForStatefulSetChanged(s.ctx, t, deploymentID, oldUpdateRevision)
	s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)
	s.crossCluster.MustWaitForStatefulSetImage(s.ctx, t, deploymentID, toSpec.ImageVersionString())
	s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, t, deploymentID)
	s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, t, deploymentID)

	// Temporary fix for MySQL - can remove after DS-4984 is completed.
	if fromSpec.DataServiceName == dataservices.MySQL {
		time.Sleep(200 * time.Second)
	}
	s.crossCluster.MustRunLoadTestJob(s.ctx, t, deploymentID)
}

func (s *PDSTestSuite) TestDataService_ScaleUp() {

	scaleNodes := map[string][]int32{
		dataservices.Cassandra:     {2, 3},
		dataservices.Consul:        {1, 3},
		dataservices.Couchbase:     {1, 2},
		dataservices.ElasticSearch: {1, 3},
		dataservices.Kafka:         {3, 5},
		dataservices.MongoDB:       {1, 2},
		dataservices.MySQL:         {1, 2},
		dataservices.Postgres:      {1, 2},
		dataservices.RabbitMQ:      {1, 3},
		dataservices.Redis:         {6, 8},
	}

	for dsName, nodeCounts := range scaleNodes {
		versions := activeVersions[dsName]
		for _, version := range versions {
			deployment := api.ShortDeploymentSpec{
				DataServiceName: dsName,
				ImageVersionTag: version,
				NodeCount:       nodeCounts[0],
			}
			scaleTo := nodeCounts[1]

			s.T().Run(fmt.Sprintf("scale-%s-%s-nodes-%v-to-%v", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount, scaleTo), func(t *testing.T) {
				t.Parallel()

				deployment.NamePrefix = fmt.Sprintf("scale-%s-", deployment.ImageVersionString())
				deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
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
				updateSpec := deployment
				updateSpec.NodeCount = scaleTo
				oldUpdateRevision := s.crossCluster.MustGetStatefulSetUpdateRevision(s.ctx, t, deploymentID)
				s.controlPlane.MustUpdateDeployment(s.ctx, t, deploymentID, &updateSpec)
				s.crossCluster.MustWaitForStatefulSetChanged(s.ctx, t, deploymentID, oldUpdateRevision)
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
}

func (s *PDSTestSuite) TestDataService_ScaleResources() {
	for dsName, versions := range activeVersions {
		for _, version := range versions {
			nodeCounts := commonNodeCounts[dsName]
			if len(nodeCounts) == 0 {
				continue
			}

			deployment := api.ShortDeploymentSpec{
				DataServiceName: dsName,
				ImageVersionTag: version,

				// Only test lowest node count.
				NodeCount: nodeCounts[0],
			}

			s.T().Run(fmt.Sprintf("scale-%s-%s-resources", deployment.DataServiceName, deployment.ImageVersionString()), func(t *testing.T) {
				t.Parallel()

				deployment.NamePrefix = fmt.Sprintf("scale-%s-", deployment.ImageVersionString())
				deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
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
				updateSpec := deployment
				updateSpec.ResourceSettingsTemplateName = dataservices.TemplateNameMed
				oldUpdateRevision := s.crossCluster.MustGetStatefulSetUpdateRevision(s.ctx, t, deploymentID)
				s.controlPlane.MustUpdateDeployment(s.ctx, t, deploymentID, &updateSpec)
				s.crossCluster.MustWaitForStatefulSetChanged(s.ctx, t, deploymentID, oldUpdateRevision)
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
}

func (s *PDSTestSuite) TestDataService_Recovery_FromDeletion() {
	for dsName, versions := range activeVersions {
		for _, version := range versions {
			nodeCounts := commonNodeCounts[dsName]
			if len(nodeCounts) == 0 {
				continue
			}

			deployment := api.ShortDeploymentSpec{
				DataServiceName: dsName,
				ImageVersionTag: version,

				// Only test highest node count.
				NodeCount: nodeCounts[len(nodeCounts)-1],
			}

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
}

func (s *PDSTestSuite) TestDataService_Metrics() {
	for dsName, versions := range activeVersions {
		for _, version := range versions {
			nodeCounts := commonNodeCounts[dsName]
			if len(nodeCounts) == 0 {
				continue
			}

			deployment := api.ShortDeploymentSpec{
				DataServiceName: dsName,
				ImageVersionTag: version,

				// Only test lowest node count.
				NodeCount: nodeCounts[0],
			}

			// MongoDB must be multi-node otherwise replication lag metrics will not be present.
			if dsName == dataservices.MongoDB {
				deployment.NodeCount = int32(2)
			}

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
}

func (s *PDSTestSuite) TestDataService_DeletePDSUser() {

	// TODO: remove this list once we have added "delete user" loadtest mode for all services.
	deleteUserServices := []string{
		dataservices.Cassandra,
		dataservices.Couchbase,
		dataservices.MongoDB,
		dataservices.MySQL,
		dataservices.Postgres,
		dataservices.RabbitMQ,
		dataservices.ElasticSearch,
	}

	for _, dsName := range deleteUserServices {
		versions := activeVersions[dsName]
		for _, version := range versions {
			nodeCounts := commonNodeCounts[dsName]
			if len(nodeCounts) == 0 {
				continue
			}
			deployment := api.ShortDeploymentSpec{
				DataServiceName: dsName,
				ImageVersionTag: version,

				// Only test lowest node count.
				NodeCount: nodeCounts[0],
			}

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

// getPatchVersionNamePrefix returns the prefix of the version name that does not change on patch update.
func getPatchVersionNamePrefix(dataServiceName, versionName string) string {
	switch dataServiceName {
	case dataservices.SqlServer:
		return "2019-CU"
	case dataservices.ElasticSearch:
		// Elasticsearch's version lines are like 8.x.y
		firstDot := strings.Index(versionName, ".")
		return versionName[0 : firstDot+1]
	default:
		// For other data services the last number is used for patch updates.
		lastDot := strings.LastIndex(versionName, ".")
		return versionName[0 : lastDot+1]
	}
}

func filterImagesByVersionNamePrefix(images []pds.ModelsImage, versionNamePrefix string) []pds.ModelsImage {
	var filteredImages []pds.ModelsImage
	for _, image := range images {
		if strings.HasPrefix(*image.Tag, versionNamePrefix) {
			filteredImages = append(filteredImages, image)
		}
	}
	return filteredImages
}

func hasPDSSystemUsersCapability(image pds.ModelsImage) bool {
	v := image.Capabilities[pdsSystemUsersCapabilityName]
	return v != nil && v != ""
}
