package copilot_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/suites/framework"
)

type CopilotTestSuite struct {
	suite.Suite
	ctx          context.Context
	ControlPlane *controlplane.ControlPlane
}

func init() {
	framework.ControlPlaneFlags()
	framework.AuthenticationFlags()
}

func TestCopilotTestSuite(t *testing.T) {
	suite.Run(t, new(CopilotTestSuite))
}

func (s *CopilotTestSuite) SetupSuite() {
	s.ctx = context.Background()
	apiClient, err := api.NewPDSClient(
		s.ctx,
		framework.PDSControlPlaneAPI,
		framework.NewLoginCredentialsFromFlags(),
	)
	s.Require().NoError(err, "could not create Control Plane API client")
	cp := framework.NewControlPlane(
		s.T(),
		apiClient,
		controlplane.WithAccountName(framework.PDSAccountName),
		controlplane.WithTenantName(framework.PDSTenantName),
		controlplane.WithProjectName(framework.PDSProjectName),
	)
	s.ControlPlane = cp
}

func (s *CopilotTestSuite) TearDownSuite() {}
