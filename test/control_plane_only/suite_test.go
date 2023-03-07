package control_plane_only

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/pds"
	"github.com/portworx/pds-integration-test/test"
)

type ControlPlaneTestSuite struct {
	suite.Suite
	ControlPlane *pds.ControlPlane
}

func TestControlPlaneTestSuite(t *testing.T) {
	suite.Run(t, new(ControlPlaneTestSuite))
}

func (s *ControlPlaneTestSuite) SetupSuite() {
	config := test.MustHaveControlPlaneEnvVariables(s.T())

	apiClient, err := pds.CreateAPIClient(config.ControlPlaneAPI)
	s.Require().NoError(err, "could not create Control Plane API client")

	actorContext, err := pds.CreateActorContextUsingApiClient(
		config.LoginCredentials, config.AccountName, config.TenantName, config.ProjectName, apiClient)
	s.Require().NoError(err, "could not create default PDS actor")

	cp := pds.NewControlPlane(apiClient, *actorContext)
	s.ControlPlane = cp
}
