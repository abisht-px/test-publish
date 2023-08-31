package targetcluster_test

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	pdsDeploymentHealthStateHealthy = "Healthy"
)

func (s *TargetClusterTestSuite) TestTargetCluster_DeletePDSChartPods() {
	postgres := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
		NodeCount:       1,
		NamePrefix:      "tc-delete-pds-chart-pods",
	}

	// Deploy a data service and wait until its healthy.
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, s.T(), &postgres)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), deploymentID)
	})

	s.MustWaitForDeploymentHealthy(deploymentID)
	s.MustWaitForDeploymentInitialized(deploymentID)
	// s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForStatefulSetReady(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForLoadBalancerServicesReady(s.ctx, s.T(), deploymentID)
	s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(s.ctx, s.T(), deploymentID)
	s.controlPlane.MustWaitForMetricsReported(s.ctx, s.T(), deploymentID)

	// Start the loadtest.
	deployment, namespace, dataServiceType := s.crossCluster.MustGetDeploymentInfo(s.ctx, s.T(), deploymentID)
	backOffLimit := pointer.Int32(0)
	job := s.crossCluster.MustCreateLoadTestJob(s.ctx, s.T(), dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), crosscluster.LoadTestCRUD, "", crosscluster.PDSUser, *deployment.NodeCount, nil, nil, backOffLimit)

	s.T().Cleanup(func() {
		err := s.targetCluster.DeleteJob(s.ctx, job.Namespace, job.Name)
		s.Require().NoError(err)
	})

	// Delete all pods in pds-system namespace and wait until they come up again.
	err := s.targetCluster.DeletePodsBySelector(s.ctx, targetcluster.PDSChartNamespace, nil)
	s.Require().NoError(err)

	s.Require().EventuallyWithT(func(t *assert.CollectT) {
		for _, operator := range targetcluster.PDSOperators {
			deployment, err := s.targetCluster.GetDeployment(s.ctx, targetcluster.PDSChartNamespace, operator.Deployment)
			assert.NoError(t, err)
			assert.Equal(t, *deployment.Spec.Replicas, deployment.Status.AvailableReplicas)
		}
	}, wait.ShortTimeout, wait.ShortRetryInterval)

	// Check the deployment loadtest succeeded.
	s.targetCluster.MustWaitForJobSuccess(s.ctx, s.T(), job.Namespace, job.Name)
	s.targetCluster.JobLogsMustNotContain(s.ctx, s.T(), job.Namespace, job.Name, "ERROR|FATAL", s.startTime)
}

func (s *TargetClusterTestSuite) MustWaitForDeploymentHealthy(deploymentID string) {
	s.T().Helper()
	s.Require().EventuallyWithT(func(t *assert.CollectT) {
		_, _, err := s.controlPlane.GetDeploymentById(s.ctx, s.T(), deploymentID)
		assert.NoError(t, err)
		deploy, _, err := s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdStatusGet(s.ctx, deploymentID).Execute()
		assert.NoError(t, err)
		healthState := deploy.GetHealth()
		assert.Equal(t, healthState, pdsDeploymentHealthStateHealthy)
	}, wait.LongTimeout, wait.LongRetryInterval)
}

func (s *TargetClusterTestSuite) MustWaitForDeploymentInitialized(deploymentID string) {
	s.T().Helper()
	s.Require().EventuallyWithT(func(t *assert.CollectT) {
		clusterInitStatus, err := s.crossCluster.GetClusterInitJob(s.ctx, s.T(), deploymentID)
		assert.NoError(t, err)
		assert.True(t, clusterInitStatus)

		nodeInitStatus, err := s.crossCluster.GetNodeInitJob(s.ctx, s.T(), deploymentID)
		assert.NoError(t, err)
		assert.True(t, nodeInitStatus)
	}, wait.LongTimeout, wait.LongRetryInterval)

}
