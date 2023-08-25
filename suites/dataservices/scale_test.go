package dataservices_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/suites/framework"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

type ScaleSuite struct {
	suite.Suite
	startTime time.Time

	controlPlane  *controlplane.ControlPlane
	targetCluster *targetcluster.TargetCluster
	crossCluster  *crosscluster.CrossClusterHelper

	activeVersions framework.DSVersionMatrix
}

func (s *ScaleSuite) SetupSuite() {
	s.startTime = time.Now()

	s.controlPlane, s.targetCluster, s.crossCluster = SetupSuite(
		s.T(),
		"ds-scale",
		controlplane.WithAccountName(framework.PDSAccountName),
		controlplane.WithTenantName(framework.PDSTenantName),
		controlplane.WithProjectName(framework.PDSProjectName),
		controlplane.WithLoadImageVersions(),
		controlplane.WithCreateTemplatesAndStorageOptions(
			framework.NewRandomName("ds-scale"),
		),
	)

	activeVersions, err := framework.NewDSVersionMatrixFromFlags()
	require.NoError(s.T(), err, "Initialize dataservices version matrix")

	s.activeVersions = activeVersions
}

func (s *ScaleSuite) TearDownSuite() {
	TearDownSuite(s.T(), s.controlPlane, s.targetCluster)
}

func (s *ScaleSuite) TestDataService_ScaleUp() {
	ctx := context.Background()

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

	for dataservice, nodeCounts := range scaleNodes {
		for _, version := range s.activeVersions.GetVersions(dataservice) {
			deployment := api.ShortDeploymentSpec{
				DataServiceName: dataservice,
				ImageVersionTag: version,
				NodeCount:       nodeCounts[0],
			}

			scaleTo := nodeCounts[1]

			testName := fmt.Sprintf(
				"%s-v%s-scale-nodes-%v-to-%v",
				deployment.DataServiceName,
				deployment.ImageVersionString(),
				deployment.NodeCount, scaleTo,
			)

			s.T().Run(testName, func(t *testing.T) {
				t.Parallel()

				deployment.NamePrefix = fmt.Sprintf("scale-%s-", deployment.ImageVersionString())
				deploymentID := s.controlPlane.MustDeployDeploymentSpec(ctx, t, &deployment)
				t.Cleanup(func() {
					s.controlPlane.MustRemoveDeployment(ctx, t, deploymentID)
					s.controlPlane.MustWaitForDeploymentRemoved(ctx, t, deploymentID)
					s.crossCluster.MustDeleteDeploymentVolumes(ctx, t, deploymentID)
				})

				// Create.
				s.controlPlane.MustWaitForDeploymentHealthy(ctx, t, deploymentID)
				s.crossCluster.MustWaitForDeploymentInitialized(ctx, t, deploymentID)
				s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx, t, deploymentID)
				s.crossCluster.MustRunLoadTestJob(ctx, t, deploymentID)

				// Update.
				updateSpec := deployment
				updateSpec.NodeCount = scaleTo
				oldUpdateRevision := s.crossCluster.MustGetStatefulSetUpdateRevision(ctx, t, deploymentID)
				s.controlPlane.MustUpdateDeployment(ctx, t, deploymentID, &updateSpec)
				s.crossCluster.MustWaitForStatefulSetChanged(ctx, t, deploymentID, oldUpdateRevision)
				s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx, t, deploymentID)

				s.crossCluster.MustRunLoadTestJob(ctx, t, deploymentID)
			})
		}
	}
}

func (s *ScaleSuite) TestDataService_ScaleResources() {
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

			testName := fmt.Sprintf(
				"scale-%s-%s-resources",
				deployment.DataServiceName,
				deployment.ImageVersionString(),
			)

			s.T().Run(testName, func(t *testing.T) {
				t.Parallel()

				deployment.NamePrefix = fmt.Sprintf("scale-%s-", deployment.ImageVersionString())
				deploymentID := s.controlPlane.MustDeployDeploymentSpec(ctx, t, &deployment)
				t.Cleanup(func() {
					s.controlPlane.MustRemoveDeployment(ctx, t, deploymentID)
					s.controlPlane.MustWaitForDeploymentRemoved(ctx, t, deploymentID)
					s.crossCluster.MustDeleteDeploymentVolumes(ctx, t, deploymentID)
				})

				// Create.
				s.controlPlane.MustWaitForDeploymentHealthy(ctx, t, deploymentID)
				s.crossCluster.MustWaitForDeploymentInitialized(ctx, t, deploymentID)
				s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx, t, deploymentID)
				s.crossCluster.MustRunLoadTestJob(ctx, t, deploymentID)

				// Update.
				updateSpec := deployment
				updateSpec.ResourceSettingsTemplateName = dataservices.TemplateNameMed
				oldUpdateRevision := s.crossCluster.MustGetStatefulSetUpdateRevision(ctx, t, deploymentID)
				s.controlPlane.MustUpdateDeployment(ctx, t, deploymentID, &updateSpec)
				s.crossCluster.MustWaitForStatefulSetChanged(ctx, t, deploymentID, oldUpdateRevision)
				s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerServicesReady(ctx, t, deploymentID)
				s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx, t, deploymentID)

				s.crossCluster.MustRunLoadTestJob(ctx, t, deploymentID)
			})
		}
	}
}
