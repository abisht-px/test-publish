package crosscluster

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	pdsModeEnvVarName = "PDS_MODE"
	pdsModeNormal     = "Normal"
)

func (c *CrossClusterHelper) MustWaitForStatefulSetReady(ctx context.Context, t tests.T, deploymentID string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, dataservices.GetLongTimeoutFor(*deployment.NodeCount), wait.RetryInterval, func(t tests.T) {
		set, err := c.targetCluster.GetStatefulSet(ctx, namespace, deployment.GetClusterResourceName())
		require.NoErrorf(t, err, "Getting statefulSet for deployment %s.", deployment.GetClusterResourceName())
		require.Equalf(t, *deployment.NodeCount, set.Status.ReadyReplicas, "ReadyReplicas don't match desired NodeCount.")
		// Also check the UpdatedReplicas count, so we are sure that all nodes are updated to the current version.
		require.Equalf(t, *deployment.NodeCount, set.Status.UpdatedReplicas, "UpdatedReplicas don't match desired NodeCount.")
	})
}

func (c *CrossClusterHelper) MustWaitForStatefulSetPDSModeNormalReady(ctx context.Context, t tests.T, deploymentID string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, dataservices.GetLongTimeoutFor(*deployment.NodeCount), wait.RetryInterval, func(t tests.T) {
		set, err := c.targetCluster.GetStatefulSet(ctx, namespace, deployment.GetClusterResourceName())
		require.NoErrorf(t, err, "Getting statefulSet for deployment %s.", deployment.GetClusterResourceName())
		pdsMode := getPDSMode(set)
		require.Containsf(t, []string{"", pdsModeNormal}, pdsMode, "PDS mode should be set to Normal")
		require.Equalf(t, *deployment.NodeCount, set.Status.ReadyReplicas, "ReadyReplicas don't match desired NodeCount.")
		// Also check the UpdatedReplicas count, so we are sure that all nodes are updated to the current version.
		require.Equalf(t, *deployment.NodeCount, set.Status.UpdatedReplicas, "UpdatedReplicas don't match desired NodeCount.")
	})
}

func (c *CrossClusterHelper) MustWaitForRestoredStatefulSetReady(ctx context.Context, t tests.T, namespace, restoreName string, nodeCount int32) {
	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		set, err := c.targetCluster.GetStatefulSet(ctx, namespace, restoreName)
		require.NoErrorf(t, err, "Getting statefulSet for deployment %s.", restoreName)
		require.Equalf(t, nodeCount, set.Status.ReadyReplicas, "ReadyReplicas don't match desired NodeCount.")
		// Also check the UpdatedReplicas count, so we are sure that all nodes are updated to the current version.
		require.Equalf(t, nodeCount, set.Status.UpdatedReplicas, "UpdatedReplicas don't match desired NodeCount.")
	})
}

func (c *CrossClusterHelper) MustGetStatefulSetUpdateRevision(ctx context.Context, t tests.T, deploymentID string) string {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()

	set, err := c.targetCluster.GetStatefulSet(ctx, namespace, deployment.GetClusterResourceName())
	require.NoErrorf(t, err, "Getting statefulSet for deployment %s.", deployment.GetClusterResourceName())
	updateRevision := set.Status.UpdateRevision
	require.NotEmpty(t, updateRevision, "UpdateRevision of the StatefulSet is empty.")
	return updateRevision
}

func (c *CrossClusterHelper) MustWaitForStatefulSetChanged(ctx context.Context, t tests.T, deploymentID, oldUpdateRevision string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		set, err := c.targetCluster.GetStatefulSet(ctx, namespace, deployment.GetClusterResourceName())
		require.NoErrorf(t, err, "Getting statefulSet for deployment %s.", deployment.GetClusterResourceName())
		updateRevision := set.Status.UpdateRevision
		require.NotEmpty(t, updateRevision, "Update revision of the StatefulSet is empty.")
		require.NotEqualf(t, oldUpdateRevision, updateRevision, "StatefulSet was not changed.")
	})
}

func (c *CrossClusterHelper) MustWaitForStatefulSetImage(ctx context.Context, t tests.T, deploymentID, imageTag string) {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := c.controlPlane.PDS.DataServicesApi.ApiDataServicesIdGet(ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
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
	case dataservices.SqlServer:
		containerName = "sqlserver"
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

func getPDSMode(set *appsv1.StatefulSet) string {
	for _, container := range set.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == pdsModeEnvVarName {
				return env.Value
			}
		}
	}
	return ""
}
