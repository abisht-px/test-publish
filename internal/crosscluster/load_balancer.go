package crosscluster

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	waiterLoadBalancerServicesReady = time.Second * 300
	waiterHostCheckFinishedTimeout  = time.Second * 60
	waiterAllHostsAvailableTimeout  = time.Second * 600
)

func (c *CrossClusterHelper) MustWaitForLoadBalancerServicesReady(ctx context.Context, t tests.T, deploymentID string) {
	deployment, resp, err := c.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, waiterLoadBalancerServicesReady, waiterRetryInterval, func(t tests.T) {
		svcs, err := c.targetCluster.ListServices(ctx, namespace, map[string]string{
			"name": deployment.GetClusterResourceName(),
		})
		require.NoErrorf(t, err, "Listing services for deployment %s.", deployment.GetClusterResourceName())

		for _, svc := range svcs.Items {
			if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
				ingress := svc.Status.LoadBalancer.Ingress
				require.NotEqualf(t, 0, len(ingress),
					"External ingress for service %s of deployment %s not assigned.",
					svc.GetClusterName(), deployment.GetClusterResourceName())
			}
		}
	})
}

func (c *CrossClusterHelper) MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx context.Context, t tests.T, deploymentID string) {
	deployment, resp, err := c.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := c.controlPlane.API.DataServicesApi.ApiDataServicesIdGet(ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)
	dataServiceType := dataService.GetName()

	if !areLoadBalancerAddressRequiredForTest(dataServiceType) {
		// Data service doesn't need load balancer addresses to be ready -> return.
		return
	}

	namespaceModel, resp, err := c.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)
	namespace := namespaceModel.GetName()

	// Collect all CNAME hostnames from DNSEndpoints.
	hostnames, err := c.targetCluster.GetDNSEndpoints(ctx, namespace, deployment.GetClusterResourceName(), "CNAME")
	require.NoError(t, err)

	// Wait until all hosts are accessible (DNS server returns an IP address for all hosts).
	if len(hostnames) > 0 {
		wait.For(t, waiterAllHostsAvailableTimeout, waiterRetryInterval, func(t tests.T) {
			dnsIPs := c.targetCluster.MustFlushDNSCache(ctx, t)
			jobNameSuffix := time.Now().Format("0405") // mmss
			jobName := c.targetCluster.MustRunHostCheckJob(ctx, t, namespace, deployment.GetClusterResourceName(), jobNameSuffix, hostnames, dnsIPs)
			c.targetCluster.MustWaitForJobSuccess(ctx, t, namespace, jobName)
		})
	}
}

func areLoadBalancerAddressRequiredForTest(dataServiceType string) bool {
	switch dataServiceType {
	case dataservices.Kafka, dataservices.RabbitMQ, dataservices.Couchbase:
		return true
	default:
		return false
	}
}
