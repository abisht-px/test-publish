package test

import (
	"fmt"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

func (s *PDSTestSuite) TestDeploymentStatuses_Available() {
	// Given.
	t := s.T()
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Cassandra,
		ImageVersionTag: "4.0.6",
		NodeCount:       1,
	}
	deployment.NamePrefix = fmt.Sprintf("initial-status-change-%s-n%d", deployment.ImageVersionString(), deployment.NodeCount)
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
	t.Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, t, deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, t, deploymentID)
	})

	// When.
	s.controlPlane.MustWaitForDeploymentManifestInitialChange(s.ctx, t, deploymentID)

	// Then.
	s.controlPlane.MustDeploymentManifestStatusHealthAvailable(s.ctx, t, deploymentID)
}
