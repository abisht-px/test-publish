package reporting_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/portworx/pds-integration-test/internal/wait"

	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

const (
	pdsDeploymentHealthStateHealthy = "Healthy"
)

func (s *ReportingTestSuite) TestDeploymentReporting_DeleteCR() {

	// Use deployment of each kind
	// TODO - Pull latest versions from CP
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
			NodeCount:       1,
			CRDNamePlural:   "postgresqls",
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Redis),
			NodeCount:       1,
			CRDNamePlural:   "redis",
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.ZooKeeper),
			NodeCount:       3,
			CRDNamePlural:   "zookeepers",
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Kafka),
			NodeCount:       1,
			CRDNamePlural:   "kafkas",
		},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("delete-cr-reporting-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			// Given
			deployment.NamePrefix = fmt.Sprintf("cr-%s-n%d", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
			})

			s.MustWaitForDeploymentHealthy(deploymentID)

			s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, t, deploymentID)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)

			// when
			s.crossCluster.MustDeleteDeploymentCustomResource(s.ctx, t, deploymentID, deployment.CRDNamePlural)

			// then
			_, resp, err := s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
			require.EqualErrorf(t, err, "404 Not Found", "Expected an error response on getting deployment %s.", deploymentID)
			require.Equalf(t, http.StatusNotFound, resp.StatusCode, "Deployment %s is not removed.", deploymentID)
		})
	}
}

func (s *ReportingTestSuite) TestDeploymentReporting_DeletetionFromCP() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
			NodeCount:       1,
			CRDNamePlural:   "postgresqls",
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Consul),
			NodeCount:       1,
			CRDNamePlural:   "consuls",
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Cassandra),
			NodeCount:       1,
			CRDNamePlural:   "cassandras",
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Redis),
			NodeCount:       1,
			CRDNamePlural:   "redis",
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.ZooKeeper),
			NodeCount:       3,
			CRDNamePlural:   "zookeepers",
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Kafka),
			NodeCount:       1,
			CRDNamePlural:   "kafkas",
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.RabbitMQ),
			NodeCount:       1,
			CRDNamePlural:   "rabbitmqs",
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.MySQL),
			NodeCount:       1,
			CRDNamePlural:   "mysqls",
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.MongoDB),
			NodeCount:       1,
			CRDNamePlural:   "mongodbs",
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.ElasticSearch),
			NodeCount:       1,
			CRDNamePlural:   "elasticsearches",
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Couchbase),
			NodeCount:       1,
			CRDNamePlural:   "couchbases",
		},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("delete-deployment-reporting-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			// Given
			deployment.NamePrefix = fmt.Sprintf("deployment-%s-n%d", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
			})

			s.MustWaitForDeploymentHealthy(deploymentID)

			s.MustWaitForDeploymentInitialized(deploymentID)
			s.crossCluster.MustWaitForStatefulSetReady(s.ctx, t, deploymentID)

			pdsDeployment, resp, err := s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
			api.RequireNoError(t, resp, err)

			namespaceModel, resp, err := s.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(s.ctx, *pdsDeployment.NamespaceId).Execute()
			api.RequireNoError(t, resp, err)
			namespace := namespaceModel.GetName()
			customResourceName := *pdsDeployment.ClusterResourceName

			// when
			s.controlPlane.MustRemoveDeployment(s.ctx, t, deploymentID)
			s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)

			// then verify CR is deleted from TC
			cr, err := s.targetCluster.GetPDSDeployment(s.ctx, namespace, deployment.CRDNamePlural, customResourceName)
			require.Nil(t, cr)

			expectedError := fmt.Sprintf("%s.deployments.pds.io %q not found", deployment.CRDNamePlural, customResourceName)
			require.EqualError(t, err, expectedError, "CR is not deleted on TC")
		})
	}
}

func (s *ReportingTestSuite) MustWaitForDeploymentHealthy(deploymentID string) {
	s.T().Helper()
	s.Require().EventuallyWithT(func(t *assert.CollectT) {
		_, _, err := s.controlPlane.GetDeploymentById(s.ctx, s.T(), deploymentID)
		assert.NoError(t, err)
		deploy, _, err := s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdStatusGet(s.ctx, deploymentID).Execute()
		assert.NoError(t, err)
		healthState := deploy.GetHealth()
		assert.Equal(t, healthState, pdsDeploymentHealthStateHealthy)
	}, wait.VeryLongTimeout, wait.LongRetryInterval)
}

func (s *ReportingTestSuite) MustWaitForDeploymentInitialized(deploymentID string) {
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
