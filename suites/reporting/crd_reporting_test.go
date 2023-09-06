package reporting_test

import (
	"context"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

// TestCRDReporting_UpdateNodeCount tests successful reporting after updating Node Count
// Steps:
// 1. Create a data service deployment and get the deployment ID
// 2. Deploy a new data service and wait for deployment to be available (or use an existing healthy one)
// 3. Observe that the statuses with number of replicas are available and correct
// 4. Edit deployment, increase the number of nodes by 1 and wait for the deployment to be available again
// 5. Observe that the statuses with number of replicas and events are updated and available
// Expected:
// 1. Deployment must be created successfully
// 2. The connection info and statuses are available as configured
// 3. Node count increased to the number specified successfully
// 4. The connection-info, statuses and events are updated successfully with the latest changes and available
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
