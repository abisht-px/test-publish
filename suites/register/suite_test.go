package register_test

import (
	"context"
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/suites/framework"
)

var (
	loginCredentials api.LoginCredentials
	apiClient        *api.PDSClient
	cp               *controlplane.ControlPlane
	tc               *targetcluster.TargetCluster

	registerOnly bool
	cleanupOnly  bool
)

type RegisterTestSuite struct {
	suite.Suite
}

func TestRegisterTestSuite(t *testing.T) {
	suite.Run(t, new(RegisterTestSuite))
}

func init() {
	framework.AuthenticationFlags()
	framework.ControlPlaneFlags()
	framework.TargetClusterFlags()

	flag.BoolVar(&registerOnly, "registerOnly", false, "Set this to true to skip cleanup")
	flag.BoolVar(&cleanupOnly, "cleanupOnly", false, "Set this to true to skip registration")
}

func (s *RegisterTestSuite) SetupSuite() {
	var err error

	loginCredentials = framework.NewLoginCredentialsFromFlags()

	apiClient, err = api.NewPDSClient(context.Background(), framework.PDSControlPlaneAPI, loginCredentials)
	require.NoError(s.T(), err, "Could not create Control Plane API client.")

	cp = framework.NewControlPlane(
		s.T(), apiClient,
		controlplane.WithAccountName(framework.PDSAccountName),
		controlplane.WithTenantName(framework.PDSTenantName),
		controlplane.WithProjectName(framework.PDSProjectName),
	)

	token := cp.MustGetServiceAccountToken(context.Background(), s.T(), framework.ServiceAccountName)
	framework.InitializePDSHelmChartVersion(s.T(), apiClient)

	tc, err = framework.NewTargetClusterFromFlags(cp.TestPDSTenantID, token)
	require.NoError(s.T(), err, "Cannot create target cluster.")
}

func (s *RegisterTestSuite) TearDownSuite() {}

func (s *RegisterTestSuite) TestRegister() {
	s.Run("Register Target Cluster", func() {
		if cleanupOnly {
			s.T().Skip("cleanupOnly flag is set to true")
		}

		framework.RegisterTargetCluster(&s.Suite, cp, tc)
	})

	s.Run("Deregister Target Cluster", func() {
		if registerOnly {
			s.T().Skip("registerOnly flag is set to true")
		}

		framework.DeregisterTargetCluster(&s.Suite, cp, tc)
	})
}
