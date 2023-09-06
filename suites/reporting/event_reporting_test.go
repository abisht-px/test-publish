package reporting_test

import (
	"context"
	"sync"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

// TestEventReporting_Successful tests successful event reporting after deploying a data service
// Steps:
// 1. Create a data service deployment and get the deployment ID
// 2. Wait for deployment to be initialized and healthy
// 3. Call GET to get the events
// Expected:
// 1. Deployment must be created successfully
// 2. Events should get reported successfully without any errors
func (s *ReportingTestSuite) TestEventReporting_Successful() {
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

	s.controlPlane.MustHaveDeploymentEventsSorted(context.Background(), s.T(), deploymentID)
	s.controlPlane.MustHaveNoDuplicateDeploymentEvents(context.Background(), s.T(), deploymentID)
}

// TestEventReporting_Update_Deployment_Successful tests successful event reporting after updating a data service
// Steps:
// 1. Create a data service deployment with 1 node and get the deployment ID
// 2. Wait for deployment to be initialized and healthy
// 3. Update the deployment to 2 nodes
// 4. Wait for deployment to be initialized and healthy
// 5. Call GET to get the events
// Expected:
// 1. Deployment must be created and updated successfully
// 2. Events should get reported successfully without any errors
func (s *ReportingTestSuite) TestEventReporting_Update_Deployment_Successful() {
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

	deployment.NodeCount = 2
	s.controlPlane.MustUpdateDeployment(context.Background(), s.T(), deploymentID, &deployment)
	s.controlPlane.MustWaitForDeploymentReplicas(context.Background(), s.T(), deploymentID, int32(deployment.NodeCount))
	s.controlPlane.MustWaitForDeploymentAvailable(context.Background(), s.T(), deploymentID)

	s.controlPlane.MustHaveDeploymentEventsSorted(context.Background(), s.T(), deploymentID)
	s.controlPlane.MustHaveNoDuplicateDeploymentEvents(context.Background(), s.T(), deploymentID)
}

// TestEventReporting_Delete_Deployment_NoEvents tests expected error on calling get events on a deleted deployment
// Steps:
// 1. Create a data service deployment and get the deployment ID
// 2. Wait for the deployment to be initialized and healthy
// 3. Delete the deployment
// 4. Wait for deployment to be deleted
// 5. Call GET to get the events
// Expected:
// 1. Deployment must be created and deleted successfully
// 2. Calling get deployment fails with errors
func (s *ReportingTestSuite) TestEventReporting_Delete_Deployment_NoEvents() {
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
		NodeCount:       1,
		NamePrefix:      dataservices.Postgres,
	}

	deploymentID := s.controlPlane.MustDeployDeploymentSpec(context.Background(), s.T(), &deployment)
	s.controlPlane.MustWaitForDeploymentAvailable(context.Background(), s.T(), deploymentID)

	s.controlPlane.MustRemoveDeploymentIfExists(context.Background(), s.T(), deploymentID)
	s.controlPlane.MustWaitForDeploymentRemoved(context.Background(), s.T(), deploymentID)

	s.controlPlane.MustGetErrorOnDeploymentEventsGet(context.Background(), s.T(), deploymentID)
}

// TestEventReporting_Failed_Deployment_No_Duplicate_Events tests duplicate events must not be reported on create
// Steps:
// 1. Create a data service deployment with 3 nodes with impossible resource allocation template and get the deployment ID
// 2. Wait for failed scheduling event to be reported
// 3. Wait for 10 minutes for the failed scheduling event to be reported again
// 4. Check if each event is reported once
// Expected:
// 1. Deployment must fail
// 2. Events should get reported successfully without any errors
// 3. Events should be reported once
func (s *ReportingTestSuite) TestEventReporting_Failed_Deployment_No_Duplicate_Events() {
	deployment := api.ShortDeploymentSpec{
		DataServiceName:              dataservices.Cassandra,
		ImageVersionTag:              "4.1.2",
		NodeCount:                    3,
		ResourceSettingsTemplateName: dataservices.TemplateNameEnormous,
	}
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, s.T(), &deployment)
	s.T().Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), deploymentID)
	})

	s.controlPlane.MustWaitForDeploymentEventCondition(s.ctx, s.T(),
		deploymentID,
		func(event pds.ModelsDeploymentTargetDeploymentEvent) bool {
			if event.Reason == nil {
				return false
			}
			reason := *event.Reason
			return reason == "FailedScheduling"
		},
		"failed pod scheduling")

	// Wait for 10 minutes and check for duplicate events
	time.Sleep(10 * time.Minute)
	s.controlPlane.MustHaveNoDuplicateDeploymentEvents(context.Background(), s.T(), deploymentID)
}

// TestEventReporting_MultipleDeployments tests successful event reporting for multiple deployments
// Steps:
// 1. Create 2 data service deployments get the deployment IDs
// 2. Wait for deployments to be initialized and healthy
// 3. Call GET to get the events
// 4. Check if events are reported for the correct deployment
// Expected:
// 1. Deployments must get created successfully
// 2. Events should get reported successfully without errors
// 3. Events should not get mixed up
func (s *ReportingTestSuite) TestEventReporting_MultipleDeployments() {
	var wg sync.WaitGroup
	var deploymentIDS []string

	deployDS := func() {
		defer wg.Done()
		// Create a new deployment.
		deployment := api.ShortDeploymentSpec{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: dsVersions.GetLatestVersion(dataservices.Postgres),
			NodeCount:       1,
			NamePrefix:      dataservices.Postgres,
		}
		deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, s.T(), &deployment)
		s.T().Cleanup(func() {
			s.controlPlane.MustRemoveDeployment(s.ctx, s.T(), deploymentID)
			s.controlPlane.MustWaitForDeploymentRemoved(s.ctx, s.T(), deploymentID)
		})
		s.controlPlane.MustWaitForDeploymentAvailable(s.ctx, s.T(), deploymentID)
		deploymentIDS = append(deploymentIDS, deploymentID)
	}

	wg.Add(2)
	go deployDS()
	go deployDS()
	wg.Wait()

	for _, deploymentID := range deploymentIDS {
		s.controlPlane.MustHaveDeploymentEventsForCorrectDeployment(context.Background(), s.T(), deploymentID)
	}
}
