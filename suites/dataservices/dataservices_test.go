package dataservices_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/kubernetes/psa"
	"github.com/portworx/pds-integration-test/internal/random"
	"github.com/portworx/pds-integration-test/internal/wait"
	"github.com/portworx/pds-integration-test/suites/framework"
)

type Dataservices struct {
	suite.Suite
	startTime time.Time

	controlPlane  *controlplane.ControlPlane
	targetCluster *targetcluster.TargetCluster
	crossCluster  *crosscluster.CrossClusterHelper

	activeVersions framework.DSVersionMatrix
}

func (s *Dataservices) SetupSuite() {
	s.startTime = time.Now()

	s.controlPlane, s.targetCluster, s.crossCluster = SetupSuite(
		s.T(),
		"ds",
		controlplane.WithAccountName(framework.PDSAccountName),
		controlplane.WithTenantName(framework.PDSTenantName),
		controlplane.WithProjectName(framework.PDSProjectName),
		controlplane.WithLoadImageVersions(),
		controlplane.WithCreateTemplatesAndStorageOptions(
			framework.NewRandomName("ds"),
		),
	)

	activeVersions, err := framework.NewDSVersionMatrixFromFlags()
	require.NoError(s.T(), err, "Initialize dataservices version matrix")

	s.activeVersions = activeVersions
}

func (s *Dataservices) TearDownSuite() {
	TearDownSuite(s.T(), s.controlPlane, s.targetCluster)
}

func (s *Dataservices) TestDataService_DeploymentWithPSA() {
	ctx := context.Background()

	for _, each := range s.activeVersions.Dataservices {
		dsName := each.Name
		versions := each.Versions

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
				_, err := s.targetCluster.CreateNamespace(ctx, namespace)
				t.Cleanup(func() {
					_ = s.targetCluster.DeleteNamespace(ctx, namespaceName)
				})
				s.Require().NoError(err)
				modelsNamespace := s.controlPlane.MustWaitForNamespaceStatus(ctx, t, namespaceName, "available")
				namespaceID := modelsNamespace.GetId()

				deployment.NamePrefix = fmt.Sprintf("deploy-%s-n%d-", deployment.ImageVersionString(), deployment.NodeCount)
				deploymentID := s.controlPlane.MustDeployDeploymentSpecIntoNamespace(ctx, t, &deployment, namespaceID)
				t.Cleanup(func() {
					s.controlPlane.MustRemoveDeployment(ctx, t, deploymentID)
					s.controlPlane.MustWaitForDeploymentRemoved(ctx, t, deploymentID)
					s.crossCluster.MustDeleteDeploymentVolumes(ctx, t, deploymentID)
				})
				s.controlPlane.MustWaitForDeploymentHealthy(ctx, t, deploymentID)
				s.crossCluster.MustWaitForDeploymentInitialized(ctx, t, deploymentID)
				s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx, t, deploymentID)

				s.crossCluster.MustRunLoadTestJob(ctx, t, deploymentID)
			})
		}
	}
}

func (s *Dataservices) TestDataService_UpdateImage() {
	ctx := context.Background()

	compatibleVersions := s.controlPlane.MustGetCompatibleVersions(ctx, s.T())
	for _, cv := range compatibleVersions {
		dataServiceName := *cv.DataServiceName

		// Filter for selected data services only.
		ok := s.activeVersions.HasDataservice(dataServiceName)
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
				s.updateTestImpl(ctx, t, fromSpec, toSpec)
			})
		}
	}
}

func (s *Dataservices) TestDataService_PDSSystemUsersV1Migration() {
	ctx := context.Background()

	dataServicesByName := s.controlPlane.MustGetDataServicesByName(ctx, s.T())
	for _, each := range s.activeVersions.Dataservices {
		dsName := each.Name
		versions := each.Versions

		if dsName == dataservices.SqlServer {
			// No need to test migration for SQL Server, as no deployments exist on Prod and Staging
			// using older images without the PDS System Users V1 feature.
			continue
		}
		dataService, ok := dataServicesByName[dsName]
		if !ok {
			assert.Fail(s.T(), "Data service with name '%s' not found", dsName)
		}
		dsImages := s.controlPlane.MustGetAllImagesForDataService(ctx, s.T(), dataService.GetId())
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

			testName := fmt.Sprintf("migrate-%s-%s-to-%s-n%d", dsName, fromSpec.ImageVersionString(), toSpec.ImageVersionString(), toSpec.NodeCount)
			s.T().Run(testName, func(t *testing.T) {
				t.Parallel()
				s.updateTestImpl(ctx, t, fromSpec, toSpec)
			})
		}
	}
}

func (s *Dataservices) TestDataService_Recovery_FromDeletion() {
	ctx := context.Background()

	for _, each := range s.activeVersions.Dataservices {
		dsName := each.Name
		versions := each.Versions

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
				deploymentID := s.controlPlane.MustDeployDeploymentSpec(ctx, t, &deployment)
				t.Cleanup(func() {
					s.controlPlane.MustRemoveDeployment(ctx, t, deploymentID)
					s.controlPlane.MustWaitForDeploymentRemoved(ctx, t, deploymentID)
					s.crossCluster.MustDeleteDeploymentVolumes(ctx, t, deploymentID)
				})
				s.controlPlane.MustWaitForDeploymentHealthy(ctx, t, deploymentID)
				s.crossCluster.MustWaitForDeploymentInitialized(ctx, t, deploymentID)
				s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx, t, deploymentID)
				s.crossCluster.MustRunLoadTestJob(ctx, t, deploymentID)
				// Delete pods and load test
				s.targetCluster.MustDeleteDeploymentPods(ctx, t, framework.TestNamespace, deploymentID)
				s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx, t, deploymentID)

				s.crossCluster.MustRunLoadTestJob(ctx, t, deploymentID)
			})
		}
	}
}

func (s *Dataservices) TestDataService_DeletePDSUser() {
	ctx := context.Background()

	for _, each := range s.activeVersions.Dataservices {
		dsName := each.Name
		versions := each.Versions

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
				deploymentID := s.controlPlane.MustDeployDeploymentSpec(ctx, t, &deployment)
				t.Cleanup(func() {
					s.controlPlane.MustRemoveDeployment(ctx, t, deploymentID)
					s.controlPlane.MustWaitForDeploymentRemoved(ctx, t, deploymentID)
					s.crossCluster.MustDeleteDeploymentVolumes(ctx, t, deploymentID)
				})
				s.controlPlane.MustWaitForDeploymentHealthy(ctx, t, deploymentID)
				s.crossCluster.MustWaitForDeploymentInitialized(ctx, t, deploymentID)
				s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx, t, deploymentID)

				// Delete 'pds' user.
				var replacePassword string
				switch deployment.DataServiceName {
				case dataservices.Consul, dataservices.Kafka:
					replacePassword = uuid.NewString()
				}
				s.crossCluster.MustRunDeleteUserJob(ctx, t, deploymentID, crosscluster.PDSUser, replacePassword)
				s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
				// Run CRUD tests with 'pds' to check that the data service fails (user does not exist).
				s.crossCluster.MustRunCRUDLoadTestJobAndFail(ctx, t, deploymentID, crosscluster.PDSUser)
				// Wait 30s before the check whether the pod was not killed due to readiness/liveness failure.
				time.Sleep(30 * time.Second)
				// Run CRUD tests with 'pds_replace_user' to check that the data service still works.
				s.crossCluster.MustRunCRUDLoadTestJob(ctx, t, deploymentID, crosscluster.PDSReplaceUser, replacePassword)
			})
		}
	}
}

func (s *Dataservices) TestDataService_ImpossibleResourceAllocation_Fails() {
	ctx := context.Background()

	deployment := api.ShortDeploymentSpec{
		DataServiceName:              dataservices.Cassandra,
		NamePrefix:                   "impossible-resources-test",
		ImageVersionTag:              "4.1.2",
		NodeCount:                    1,
		ResourceSettingsTemplateName: dataservices.TemplateNameEnormous,
	}
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(ctx, s.T(), deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(ctx, s.T(), deploymentID)
		s.crossCluster.MustDeleteDeploymentVolumes(ctx, s.T(), deploymentID)
	})

	// Wait for the standard timeout, and then make sure the deployment is unavailable.
	time.Sleep(wait.StandardTimeout)
	s.controlPlane.MustDeploymentManifestStatusHealthUnavailable(ctx, s.T(), deploymentID)
}
