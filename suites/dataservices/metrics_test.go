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

type MetricsSuite struct {
	suite.Suite
	startTime time.Time

	controlPlane  *controlplane.ControlPlane
	targetCluster *targetcluster.TargetCluster
	crossCluster  *crosscluster.CrossClusterHelper

	activeVersions framework.DSVersionMatrix
}

func (s *MetricsSuite) SetupSuite() {
	s.startTime = time.Now()

	s.controlPlane, s.targetCluster, s.crossCluster = SetupSuite(
		s.T(),
		"ds-metrics",
		controlplane.WithAccountName(framework.PDSAccountName),
		controlplane.WithTenantName(framework.PDSTenantName),
		controlplane.WithProjectName(framework.PDSProjectName),
		controlplane.WithLoadImageVersions(),
		controlplane.WithCreateTemplatesAndStorageOptions(
			framework.NewRandomName("ds-metrics"),
		),
		controlplane.WithPrometheusClient(
			framework.PDSControlPlaneAPI,
			framework.NewLoginCredentialsFromFlags(),
		),
	)

	activeVersions, err := framework.NewDSVersionMatrixFromFlags()
	require.NoError(s.T(), err, "Initialize dataservices version matrix")

	s.activeVersions = activeVersions
}

func (s *MetricsSuite) TearDownSuite() {
	TearDownSuite(s.T(), s.controlPlane, s.targetCluster)
}

func (s *MetricsSuite) TestDataService_Metrics() {
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

			// MongoDB must be multi-node otherwise replication lag metrics will not be present.
			if dsName == dataservices.MongoDB {
				deployment.NodeCount = int32(2)
			}

			s.T().Run(fmt.Sprintf("metrics-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
				t.Parallel()

				deployment.NamePrefix = fmt.Sprintf("metrics-%s-n%d-", deployment.ImageVersionString(), deployment.NodeCount)
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

				// Try to get DS metrics from prometheus.
				s.controlPlane.MustWaitForMetricsReported(ctx, t, deploymentID)
			})
		}
	}
}
