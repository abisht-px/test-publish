package crosscluster

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	waiterStatefulSetReadyAndUpdatedReplicas = time.Minute * 10
)

func (c *CrossClusterHelper) MustEnsureStatefulSetReadyAndUpdatedReplicas(ctx context.Context, t tests.T, deploymentID string) {
	deployment, resp, err := c.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, waiterStatefulSetReadyAndUpdatedReplicas, waiterRetryInterval, func(t tests.T) {
		set, err := c.targetCluster.GetStatefulSet(ctx, namespace, deployment.GetClusterResourceName())
		require.NoErrorf(t, err, "Getting statefulSet for deployment %s.", deployment.GetClusterResourceName())
		require.Equalf(t, *deployment.NodeCount, set.Status.ReadyReplicas, "ReadyReplicas don't match desired NodeCount.")
		// Also check the UpdatedReplicas count, so we are sure that all nodes were restarted after the change.
		require.Equalf(t, *deployment.NodeCount, set.Status.UpdatedReplicas, "UpdatedReplicas don't match desired NodeCount.")
	})
}

func (c *CrossClusterHelper) MustEnsureStatefulSetImage(ctx context.Context, t tests.T, deploymentID, imageTag string) {
	deployment, resp, err := c.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := c.controlPlane.API.DataServicesApi.ApiDataServicesIdGet(ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, waiterDeploymentStatusHealthyTimeout, waiterRetryInterval, func(t tests.T) {
		set, err := c.targetCluster.GetStatefulSet(ctx, namespace, deployment.GetClusterResourceName())
		require.NoErrorf(t, err, "Getting statefulSet for deployment %s.", deployment.GetClusterResourceName())

		image, err := getDatabaseImage(dataService.GetName(), set)
		require.NoErrorf(t, err, "Getting database image of deployment %s.", deployment.GetClusterResourceName())

		require.Contains(t, image, imageTag, "StatefulSet %s does not contain image tag %q.", deployment.GetClusterResourceName(), imageTag)
	})
}

func getDatabaseImage(deploymentType string, set *appsv1.StatefulSet) (string, error) {
	var containerName string
	switch deploymentType {
	case dataservices.Postgres:
		containerName = "postgresql"
	case dataservices.Cassandra:
		containerName = "cassandra"
	case dataservices.Couchbase:
		containerName = "couchbase"
	case dataservices.Redis:
		containerName = "redis"
	case dataservices.ZooKeeper:
		containerName = "zookeeper"
	case dataservices.Kafka:
		containerName = "kafka"
	case dataservices.RabbitMQ:
		containerName = "rabbitmq"
	case dataservices.MongoDB:
		containerName = "mongos"
	case dataservices.MySQL:
		containerName = "mysql"
	case dataservices.ElasticSearch:
		containerName = "elasticsearch"
	case dataservices.Consul:
		containerName = "consul"
	default:
		return "", fmt.Errorf("unknown database type: %s", deploymentType)
	}

	for _, container := range set.Spec.Template.Spec.Containers {
		if container.Name != containerName {
			continue
		}

		return container.Image, nil
	}

	return "", fmt.Errorf("database type: %s: container %q is not found", deploymentType, containerName)
}
