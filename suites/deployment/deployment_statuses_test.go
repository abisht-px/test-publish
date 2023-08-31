package deployment_test

import (
	"fmt"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

func (s *DeploymentTestSuite) TestDeploymentStatuses_Available() {
	// Given.
	t := s.T()
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Cassandra),
		NodeCount:       1,
	}
	deployment.NamePrefix = fmt.Sprintf("initial-status-change-%s-n%d", deployment.ImageVersionString(), deployment.NodeCount)
	deploymentID := controlPlane.MustDeployDeploymentSpec(ctx, t, &deployment)
	t.Cleanup(func() {
		controlPlane.MustRemoveDeployment(ctx, t, deploymentID)
		controlPlane.MustWaitForDeploymentRemoved(ctx, t, deploymentID)
	})

	// When.
	controlPlane.MustWaitForDeploymentManifestInitialChange(ctx, t, deploymentID)

	// Then.
	controlPlane.MustDeploymentManifestStatusHealthAvailable(ctx, t, deploymentID)
}
