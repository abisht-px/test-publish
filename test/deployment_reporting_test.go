package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

func (s *PDSTestSuite) TestDeploymentReporting_DeleteCR() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.6",
			NodeCount:       1,
			CRDNamePlural:   "postgresqls",
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.14.0",
			NodeCount:       1,
			CRDNamePlural:   "consuls",
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.6",
			NodeCount:       1,
			CRDNamePlural:   "cassandras",
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.5",
			NodeCount:       1,
			CRDNamePlural:   "redis",
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.7.1",
			NodeCount:       3,
			CRDNamePlural:   "zookeepers",
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.8.0",
			NodeCount:       3,
			CRDNamePlural:   "zookeepers",
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.1.1",
			NodeCount:       1,
			CRDNamePlural:   "kafkas",
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       1,
			CRDNamePlural:   "kafkas",
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.10.9",
			NodeCount:       1,
			CRDNamePlural:   "rabbitmqs",
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.31",
			NodeCount:       1,
			CRDNamePlural:   "mysqls",
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.3",
			NodeCount:       1,
			CRDNamePlural:   "mongodbs",
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.5.2",
			NodeCount:       1,
			CRDNamePlural:   "elasticsearches",
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       1,
			CRDNamePlural:   "couchbases",
		},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("delete-cr-reporting-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			// Given
			deployment.NamePrefix = fmt.Sprintf("delete-cr-reporting-%s-n%d", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
			})
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)

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

func (s *PDSTestSuite) TestDeploymentReporting_DeletetionFromCP() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: "14.6",
			NodeCount:       1,
			CRDNamePlural:   "postgresqls",
		},
		{
			DataServiceName: dataservices.Consul,
			ImageVersionTag: "1.14.0",
			NodeCount:       1,
			CRDNamePlural:   "consuls",
		},
		{
			DataServiceName: dataservices.Cassandra,
			ImageVersionTag: "4.0.6",
			NodeCount:       1,
			CRDNamePlural:   "cassandras",
		},
		{
			DataServiceName: dataservices.Redis,
			ImageVersionTag: "7.0.5",
			NodeCount:       1,
			CRDNamePlural:   "redis",
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.7.1",
			NodeCount:       3,
			CRDNamePlural:   "zookeepers",
		},
		{
			DataServiceName: dataservices.ZooKeeper,
			ImageVersionTag: "3.8.0",
			NodeCount:       3,
			CRDNamePlural:   "zookeepers",
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.1.1",
			NodeCount:       1,
			CRDNamePlural:   "kafkas",
		},
		{
			DataServiceName: dataservices.Kafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       1,
			CRDNamePlural:   "kafkas",
		},
		{
			DataServiceName: dataservices.RabbitMQ,
			ImageVersionTag: "3.10.9",
			NodeCount:       1,
			CRDNamePlural:   "rabbitmqs",
		},
		{
			DataServiceName: dataservices.MySQL,
			ImageVersionTag: "8.0.31",
			NodeCount:       1,
			CRDNamePlural:   "mysqls",
		},
		{
			DataServiceName: dataservices.MongoDB,
			ImageVersionTag: "6.0.3",
			NodeCount:       1,
			CRDNamePlural:   "mongodbs",
		},
		{
			DataServiceName: dataservices.ElasticSearch,
			ImageVersionTag: "8.5.2",
			NodeCount:       1,
			CRDNamePlural:   "elasticsearches",
		},
		{
			DataServiceName: dataservices.Couchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       1,
			CRDNamePlural:   "couchbases",
		},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("delete-deployment-reporting-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			// Given
			deployment.NamePrefix = fmt.Sprintf("delete-deployment-reporting-%s-n%d", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
			})
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)

			s.crossCluster.MustWaitForDeploymentInitialized(s.ctx, t, deploymentID)
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
