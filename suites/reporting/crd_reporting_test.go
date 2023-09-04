package reporting_test

import (
	"context"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

func (s *ReportingTestSuite) TestCRDReporting_UpdateNodeCount() {
	// Create a new deployment.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
		NodeCount:       1,
		NamePrefix:      dataservices.Postgres,
	}

	deploymentID := s.controlPlane.MustDeployDeploymentSpec(context.Background(), s.T(), &deployment)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(context.Background(), s.T(), deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(context.Background(), s.T(), deploymentID)
	})
	s.controlPlane.MustWaitForDeploymentAvailable(context.Background(), s.T(), deploymentID)

	// Update the node count of the deployment.
	deployment.NodeCount = 2
	s.controlPlane.MustUpdateDeployment(context.Background(), s.T(), deploymentID, &deployment)
	s.controlPlane.MustWaitForDeploymentReplicas(context.Background(), s.T(), deploymentID, int32(deployment.NodeCount))
	s.controlPlane.MustWaitForDeploymentAvailable(context.Background(), s.T(), deploymentID)
}