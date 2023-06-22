package crosscluster

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
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

func (c *CrossClusterHelper) mustGetDeploymentInfo(ctx context.Context, t *testing.T, deploymentID string) (*pds.ModelsDeployment, *pds.ModelsNamespace, string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)
	namespace, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)
	dataService, resp, err := c.controlPlane.PDS.DataServicesApi.ApiDataServicesIdGet(ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)
	dataServiceType := dataService.GetName()
	return deployment, namespace, dataServiceType
}

func (c *CrossClusterHelper) MustRunLoadTestJob(ctx context.Context, t *testing.T, deploymentID string) {
	deployment, namespace, dataServiceType := c.mustGetDeploymentInfo(ctx, t, deploymentID)
	c.MustRunGenericLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestCRUD, "", PDSUser, *deployment.NodeCount, nil)
}

func (c *CrossClusterHelper) MustRunReadLoadTestJob(ctx context.Context, t *testing.T, deploymentID, seed string) {
	deployment, namespace, dataServiceType := c.mustGetDeploymentInfo(ctx, t, deploymentID)
	c.MustRunGenericLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestRead, seed, PDSUser, *deployment.NodeCount, nil)
}

func (c *CrossClusterHelper) MustRunWriteLoadTestJob(ctx context.Context, t *testing.T, deploymentID, seed string) {
	deployment, namespace, dataServiceType := c.mustGetDeploymentInfo(ctx, t, deploymentID)
	c.MustRunGenericLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestWrite, seed, PDSUser, *deployment.NodeCount, nil)
}

func (c *CrossClusterHelper) MustRunCRUDLoadTestJob(ctx context.Context, t *testing.T, deploymentID, user string) {
	deployment, namespace, dataServiceType := c.mustGetDeploymentInfo(ctx, t, deploymentID)
	c.MustRunGenericLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestCRUD, "", user, *deployment.NodeCount, nil)
}

func (c *CrossClusterHelper) MustRunCRUDLoadTestJobAndFail(ctx context.Context, t *testing.T, deploymentID, user string) {
	deployment, namespace, dataServiceType := c.mustGetDeploymentInfo(ctx, t, deploymentID)
	jobNamespace, jobName := c.mustCreateLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestCRUD, "", user, *deployment.NodeCount, nil)
	c.targetCluster.MustWaitForLoadTestFailure(ctx, t, jobNamespace, jobName)
}

func (c *CrossClusterHelper) MustRunDeleteUserJob(ctx context.Context, t *testing.T, deploymentID, user string) {
	extraEnv := map[string]string{
		"DELETE_USER": user,
	}
	deployment, namespace, dataServiceType := c.mustGetDeploymentInfo(ctx, t, deploymentID)
	c.MustRunGenericLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestDeleteUser, "", user, *deployment.NodeCount, extraEnv)
}

func (c *CrossClusterHelper) MustRunGenericLoadTestJob(ctx context.Context, t *testing.T, dataServiceType, namespace, deploymentName, mode, seed, user string, nodeCount int32, extraEnv map[string]string) {
	jobNamespace, jobName := c.mustCreateLoadTestJob(ctx, t, dataServiceType, namespace, deploymentName, mode, seed, user, nodeCount, extraEnv)
	c.targetCluster.MustWaitForLoadTestSuccess(ctx, t, jobNamespace, jobName, c.startTime)
	c.targetCluster.JobLogsMustNotContain(ctx, t, jobNamespace, jobName, "ERROR|FATAL", c.startTime)
}

func (c *CrossClusterHelper) mustCreateLoadTestJob(ctx context.Context, t *testing.T, dataServiceType, namespace, deploymentName, mode, seed, user string, nodeCount int32, extraEnv map[string]string) (string, string) {

	jobName := fmt.Sprintf("%s-loadtest-%d", deploymentName, time.Now().Unix())
	if mode != "" {
		suffix := strings.ReplaceAll(mode, "_", "")
		jobName = fmt.Sprintf("%s-loadtest-%s-%d", deploymentName, suffix, time.Now().Unix())
	}

	image, err := getLoadTestJobImage(dataServiceType)
	require.NoError(t, err)

	env := c.targetCluster.MustGetLoadTestJobEnv(ctx, t, dataServiceType, deploymentName, namespace, mode, seed, user, nodeCount, extraEnv)

	job, err := c.targetCluster.CreateJob(ctx, namespace, jobName, image, env, nil)
	require.NoError(t, err)

	return namespace, job.GetName()
}

func getLoadTestJobImage(dataServiceType string) (string, error) {
	image, ok := loadTestImages[dataServiceType]
	if !ok {
		return "", fmt.Errorf("loadtest job image not found for data service %s", dataServiceType)
	}
	return image, nil
}
