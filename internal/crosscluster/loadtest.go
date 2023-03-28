package crosscluster

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

var loadTestImages = map[string]string{
	dataservices.Cassandra:     "portworx/pds-loadtests:cassandra-0.0.5",
	dataservices.Couchbase:     "portworx/pds-loadtests:couchbase-0.0.3",
	dataservices.Redis:         "portworx/pds-loadtests:redis-0.0.3",
	dataservices.ZooKeeper:     "portworx/pds-loadtests:zookeeper-0.0.2",
	dataservices.Kafka:         "portworx/pds-loadtests:kafka-0.0.3",
	dataservices.RabbitMQ:      "portworx/pds-loadtests:rabbitmq-0.0.2",
	dataservices.MongoDB:       "portworx/pds-loadtests:mongodb-0.0.1",
	dataservices.MySQL:         "portworx/pds-loadtests:mysql-0.0.3",
	dataservices.ElasticSearch: "portworx/pds-loadtests:elasticsearch-0.0.2",
	dataservices.Consul:        "portworx/pds-loadtests:consul-0.0.1",
	dataservices.Postgres:      "portworx/pds-loadtests:postgresql-0.0.3",
}

func (c *CrossClusterHelper) MustRunLoadTestJob(ctx context.Context, t *testing.T, deploymentID string) {
	jobNamespace, jobName := c.mustCreateLoadTestJob(ctx, t, deploymentID)
	c.targetCluster.MustWaitForLoadTestSuccess(ctx, t, jobNamespace, jobName, c.startTime)
	c.targetCluster.JobLogsMustNotContain(ctx, t, jobNamespace, jobName, "ERROR|FATAL", c.startTime)
}

func (c *CrossClusterHelper) mustCreateLoadTestJob(ctx context.Context, t *testing.T, deploymentID string) (string, string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)
	deploymentName := deployment.GetClusterResourceName()

	namespace, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := c.controlPlane.PDS.DataServicesApi.ApiDataServicesIdGet(ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)
	dataServiceType := dataService.GetName()

	dsImage, resp, err := c.controlPlane.PDS.ImagesApi.ApiImagesIdGet(ctx, deployment.GetImageId()).Execute()
	api.RequireNoError(t, resp, err)
	dsImageCreatedAt := dsImage.GetCreatedAt()

	jobName := fmt.Sprintf("%s-loadtest-%d", deployment.GetClusterResourceName(), time.Now().Unix())

	image, err := getLoadTestJobImage(dataServiceType)
	require.NoError(t, err)

	env := c.targetCluster.MustGetLoadTestJobEnv(ctx, t, dataService, dsImageCreatedAt, deploymentName, namespace.GetName(), deployment.NodeCount)

	job, err := c.targetCluster.CreateJob(ctx, namespace.GetName(), jobName, image, env, nil)
	require.NoError(t, err)

	return namespace.GetName(), job.GetName()
}

func getLoadTestJobImage(dataServiceType string) (string, error) {
	image, ok := loadTestImages[dataServiceType]
	if !ok {
		return "", fmt.Errorf("loadtest job image not found for data service %s", dataServiceType)
	}
	return image, nil
}
