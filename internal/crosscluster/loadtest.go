package crosscluster

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

const (
	LoadTestRead       = "read"
	LoadTestWrite      = "write"
	LoadTestCRUD       = "crud"
	LoadTestDeleteUser = "delete_user"

	PDSUser        = "pds"
	PDSReplaceUser = "pds_replace_user"

	DefaultLoadTestImage = "portworx/pds-loadtests:sample-load-0.1.0"
)

var loadTestImages = map[string]string{
	dataservices.Cassandra:     DefaultLoadTestImage,
	dataservices.Couchbase:     DefaultLoadTestImage,
	dataservices.Redis:         DefaultLoadTestImage,
	dataservices.ZooKeeper:     DefaultLoadTestImage,
	dataservices.Kafka:         DefaultLoadTestImage,
	dataservices.RabbitMQ:      DefaultLoadTestImage,
	dataservices.MongoDB:       DefaultLoadTestImage,
	dataservices.MySQL:         DefaultLoadTestImage,
	dataservices.ElasticSearch: DefaultLoadTestImage,
	dataservices.Consul:        DefaultLoadTestImage,
	dataservices.Postgres:      DefaultLoadTestImage,
	dataservices.SqlServer:     DefaultLoadTestImage,
}

func (c *CrossClusterHelper) MustRunLoadTestJob(ctx context.Context, t *testing.T, deploymentID string) {
	c.mustRunGenericLoadTestJob(ctx, t, deploymentID, "", LoadTestCRUD, "", PDSUser, nil)
}

func (c *CrossClusterHelper) MustRunReadLoadTestJob(ctx context.Context, t *testing.T, deploymentID, clusterResourceName, seed string) {
	c.mustRunGenericLoadTestJob(ctx, t, deploymentID, clusterResourceName, LoadTestRead, seed, PDSUser, nil)
}

func (c *CrossClusterHelper) MustRunWriteLoadTestJob(ctx context.Context, t *testing.T, deploymentID, clusterResourceName, seed string) {
	c.mustRunGenericLoadTestJob(ctx, t, deploymentID, clusterResourceName, LoadTestWrite, seed, PDSUser, nil)
}

func (c *CrossClusterHelper) MustRunCRUDLoadTestJob(ctx context.Context, t *testing.T, deploymentID, clusterResourceName, user string) {
	c.mustRunGenericLoadTestJob(ctx, t, deploymentID, clusterResourceName, LoadTestCRUD, "", user, nil)
}

func (c *CrossClusterHelper) MustRunCRUDLoadTestJobAndFail(ctx context.Context, t *testing.T, deploymentID, clusterResourceName, user string) {
	jobNamespace, jobName := c.mustCreateLoadTestJob(ctx, t, deploymentID, clusterResourceName, LoadTestCRUD, "", user, nil)
	c.targetCluster.MustWaitForLoadTestFailure(ctx, t, jobNamespace, jobName, c.startTime)
}

func (c *CrossClusterHelper) MustRunDeleteUserJob(ctx context.Context, t *testing.T, deploymentID, user string) {
	extraEnv := map[string]string{
		"DELETE_USER": user,
	}
	c.mustRunGenericLoadTestJob(ctx, t, deploymentID, "", LoadTestDeleteUser, "", user, extraEnv)
}

func (c *CrossClusterHelper) mustRunGenericLoadTestJob(ctx context.Context, t *testing.T, deploymentID, clusterResourceName, mode, seed, user string, extraEnv map[string]string) {
	jobNamespace, jobName := c.mustCreateLoadTestJob(ctx, t, deploymentID, clusterResourceName, mode, seed, user, extraEnv)
	c.targetCluster.MustWaitForLoadTestSuccess(ctx, t, jobNamespace, jobName, c.startTime)
	c.targetCluster.JobLogsMustNotContain(ctx, t, jobNamespace, jobName, "ERROR|FATAL", c.startTime)
}

func (c *CrossClusterHelper) mustCreateLoadTestJob(ctx context.Context, t *testing.T, deploymentID, clusterResourceName, mode, seed, user string, extraEnv map[string]string) (string, string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)
	deploymentName := deployment.GetClusterResourceName()
	if clusterResourceName != "" {
		deploymentName = clusterResourceName
	}

	namespace, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := c.controlPlane.PDS.DataServicesApi.ApiDataServicesIdGet(ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)
	dataServiceType := dataService.GetName()

	dsImage, resp, err := c.controlPlane.PDS.ImagesApi.ApiImagesIdGet(ctx, deployment.GetImageId()).Execute()
	api.RequireNoError(t, resp, err)
	dsImageCreatedAt := dsImage.GetCreatedAt()

	jobName := fmt.Sprintf("%s-loadtest-%d", deploymentName, time.Now().Unix())
	if mode != "" {
		suffix := strings.ReplaceAll(mode, "_", "")
		jobName = fmt.Sprintf("%s-loadtest-%s-%d", deploymentName, suffix, time.Now().Unix())
	}

	image, err := getLoadTestJobImage(dataServiceType)
	require.NoError(t, err)

	env := c.targetCluster.MustGetLoadTestJobEnv(ctx, t, dataService, dsImageCreatedAt, deploymentName, namespace.GetName(), mode, seed, user, deployment.NodeCount, extraEnv)

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
