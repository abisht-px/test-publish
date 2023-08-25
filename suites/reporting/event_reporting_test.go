package reporting_test

import (
	"context"
	"sync"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
)

const PostgreSQLImageTag = "15.3"

func (s *ReportingTestSuite) TestEventReporting_Successful() {
	// Create a new deployment.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: PostgreSQLImageTag,
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

func (s *ReportingTestSuite) TestEventReporting_Update_Deployment_Successful() {
	// Create a new deployment.
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: PostgreSQLImageTag,
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

func (s *ReportingTestSuite) TestEventReporting_Delete_Deployment_NoEvents() {
	deployment := api.ShortDeploymentSpec{
		DataServiceName: dataservices.Postgres,
		ImageVersionTag: PostgreSQLImageTag,
		NodeCount:       1,
		NamePrefix:      dataservices.Postgres,
	}

	deploymentID := s.controlPlane.MustDeployDeploymentSpec(context.Background(), s.T(), &deployment)
	s.controlPlane.MustWaitForDeploymentAvailable(context.Background(), s.T(), deploymentID)

	s.controlPlane.MustRemoveDeploymentIfExists(context.Background(), s.T(), deploymentID)
	s.controlPlane.MustWaitForDeploymentRemoved(context.Background(), s.T(), deploymentID)

	s.controlPlane.MustGetErrorOnDeploymentEventsGet(context.Background(), s.T(), deploymentID)
}

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

func (s *ReportingTestSuite) TestEventReporting_MultipleDeployments() {
	var wg sync.WaitGroup
	var deploymentIDS []string

	deployDS := func() {
		defer wg.Done()
		// Create a new deployment.
		deployment := api.ShortDeploymentSpec{
			DataServiceName: dataservices.Postgres,
			ImageVersionTag: PostgreSQLImageTag,
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
