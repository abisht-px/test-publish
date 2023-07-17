package crosscluster

import (
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *CrossClusterHelper) MustHaveDeploymentsMatching(ctx context.Context, t tests.T, deploymentID string) {
	eventsResponse, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdEventsGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	m := make(map[string]bool)
	for _, e := range eventsResponse {
		m[*e.Name] = true
	}

	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)
	namespace := namespaceModel.GetName()

	customResourceName := *deployment.ClusterResourceName

	db, err := c.targetCluster.GetPDSDatabase(ctx, namespace, customResourceName)
	require.NoErrorf(t, err, "Getting database %s from target cluster failed", customResourceName)

	for _, e := range db.Status.ResourceEvents {
		for _, evt := range e.Events {
			if _, ok := m[evt.Name]; !ok {
				assert.Truef(t, ok, "No event %s found in get deployments event response", evt.Name)
			}
		}
	}
}
