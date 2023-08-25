package iam_test

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/suites/framework"
)

var (
	authUserName string
	authUserPwd  string
)

type IAMTestSuite struct {
	suite.Suite
	ctx          context.Context
	ControlPlane *controlplane.ControlPlane
}

func init() {
	framework.ControlPlaneFlags()
	framework.AuthenticationFlags()

	flag.StringVar(&authUserName, "authUserName", "pds-test-auth-user@purestorage.com", "Auth User Name (pds-test-auth-user@purestorage.com)")
	flag.StringVar(&authUserPwd, "authPassword", "", "Auth User Password for pds-test-auth-user@purestorage.com")
}

func TestIAMTestSuite(t *testing.T) {
	suite.Run(t, new(IAMTestSuite))
}

func (s *IAMTestSuite) SetupSuite() {
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

func (s *IAMTestSuite) TearDownSuite() {}
