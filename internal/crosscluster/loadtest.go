package crosscluster

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/utils/pointer"

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

	DefaultLoadTestImage = "portworx/pds-loadtests:sample-load-0.1.8"
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

func (c *CrossClusterHelper) MustGetDeploymentInfo(ctx context.Context, t *testing.T, deploymentID string) (*pds.ModelsDeployment, *pds.ModelsNamespace, string) {
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
	deployment, namespace, dataServiceType := c.MustGetDeploymentInfo(ctx, t, deploymentID)
	user := PDSUser
	if dataServiceType == dataservices.Redis {
		dsImage, resp, err := c.controlPlane.PDS.ImagesApi.ApiImagesIdGet(ctx, deployment.GetImageId()).Execute()
		api.RequireNoError(t, resp, err)
		if *dsImage.Tag < "7.0.5" {
			// Older images before this change: https://github.com/portworx/pds-images-redis/pull/61 had "default" user.
			user = "default"
		}
	}
	c.MustRunGenericLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestCRUD, "", user, *deployment.NodeCount, nil)
}

func (c *CrossClusterHelper) MustRunReadLoadTestJob(ctx context.Context, t *testing.T, deploymentID, seed string) {
	deployment, namespace, dataServiceType := c.MustGetDeploymentInfo(ctx, t, deploymentID)
	c.MustRunGenericLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestRead, seed, PDSUser, *deployment.NodeCount, nil)
}

func (c *CrossClusterHelper) MustRunWriteLoadTestJob(ctx context.Context, t *testing.T, deploymentID, seed string) {
	deployment, namespace, dataServiceType := c.MustGetDeploymentInfo(ctx, t, deploymentID)
	c.MustRunGenericLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestWrite, seed, PDSUser, *deployment.NodeCount, nil)
}

func (c *CrossClusterHelper) MustRunCRUDLoadTestJob(ctx context.Context, t *testing.T, deploymentID, user, replaceToken string) {
	deployment, namespace, dataServiceType := c.MustGetDeploymentInfo(ctx, t, deploymentID)
	var extraEnv map[string]string
	if replaceToken != "" {
		extraEnv = map[string]string{
			"PASSWORD": replaceToken,
		}
	}
	c.MustRunGenericLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestCRUD, "", user, *deployment.NodeCount, extraEnv)
}

func (c *CrossClusterHelper) MustRunCRUDLoadTestJobAndFail(ctx context.Context, t *testing.T, deploymentID, user string) {
	deployment, namespace, dataServiceType := c.MustGetDeploymentInfo(ctx, t, deploymentID)
	job := c.MustCreateLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestCRUD, "", user, *deployment.NodeCount, nil, pointer.Int32(30))
	c.targetCluster.MustWaitForLoadTestFailure(ctx, t, job.Namespace, job.Name)
}

func (c *CrossClusterHelper) MustRunDeleteUserJob(ctx context.Context, t *testing.T, deploymentID, user, replacePassword string) {
	extraEnv := map[string]string{
		"DELETE_USER": user,
	}
	if replacePassword != "" {
		extraEnv["REPLACE_PASSWORD"] = replacePassword
	}
	deployment, namespace, dataServiceType := c.MustGetDeploymentInfo(ctx, t, deploymentID)
	c.MustRunGenericLoadTestJob(ctx, t, dataServiceType, namespace.GetName(), deployment.GetClusterResourceName(), LoadTestDeleteUser, "", user, *deployment.NodeCount, extraEnv)
}

func (c *CrossClusterHelper) MustRunGenericLoadTestJob(ctx context.Context, t *testing.T, dataServiceType, namespace, deploymentName, mode, seed, user string, nodeCount int32, extraEnv map[string]string) {
	job := c.MustCreateLoadTestJob(ctx, t, dataServiceType, namespace, deploymentName, mode, seed, user, nodeCount, extraEnv, pointer.Int32(30))
	c.targetCluster.MustWaitForLoadTestSuccess(ctx, t, job.Namespace, job.Name, c.startTime)
	c.targetCluster.JobLogsMustNotContain(ctx, t, job.Namespace, job.Name, "ERROR|FATAL", c.startTime)
}

func (c *CrossClusterHelper) MustCreateLoadTestJob(ctx context.Context, t *testing.T, dataServiceType, namespace, deploymentName, mode, seed, user string, nodeCount int32, extraEnv map[string]string, ttlSecondsAfterFinished *int32) *batchv1.Job {

	jobName := fmt.Sprintf("%s-loadtest-%d", deploymentName, time.Now().Unix())
	if mode != "" {
		suffix := strings.ReplaceAll(mode, "_", "")
		jobName = fmt.Sprintf("%s-loadtest-%s-%d", deploymentName, suffix, time.Now().Unix())
	}

	image, err := getLoadTestJobImage(dataServiceType)
	require.NoError(t, err)

	env := c.targetCluster.MustGetLoadTestJobEnv(ctx, t, dataServiceType, deploymentName, namespace, mode, seed, user, nodeCount, extraEnv)

	job, err := c.targetCluster.CreateJob(ctx, namespace, jobName, image, env, nil, ttlSecondsAfterFinished)
	require.NoError(t, err)

	return job
}

func getLoadTestJobImage(dataServiceType string) (string, error) {
	image, ok := loadTestImages[dataServiceType]
	if !ok {
		return "", fmt.Errorf("loadtest job image not found for data service %s", dataServiceType)
	}
	return image, nil
}
