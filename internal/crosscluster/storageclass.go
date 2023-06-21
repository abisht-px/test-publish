package crosscluster

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/require"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *CrossClusterHelper) MustGetStorageClassesForDeployment(ctx context.Context, t tests.T, deploymentID string) []*storagev1.StorageClass {
	deployment, resp, err := c.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := c.controlPlane.PDS.NamespacesApi.ApiNamespacesIdGet(ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()

	var storageClasses []*storagev1.StorageClass
	storageClassNames := []string{
		fmt.Sprintf("%s-%s", deployment.GetClusterResourceName(), namespace),
		fmt.Sprintf("%s-sharedbackups-%s", deployment.GetClusterResourceName(), namespace),
	}
	for _, name := range storageClassNames {
		sc, err := c.targetCluster.GetStorageClass(ctx, name)
		require.NoError(t, err)
		storageClasses = append(storageClasses, sc)
	}
	return storageClasses
}
