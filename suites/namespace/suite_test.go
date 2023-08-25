package namespace_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/suites/framework"
)

var loginCredentials api.LoginCredentials

type NamespaceTestSuite struct {
	suite.Suite
	ctx       context.Context
	startTime time.Time

	controlPlane  *controlplane.ControlPlane
	targetCluster *targetcluster.TargetCluster
	crossCluster  *crosscluster.CrossClusterHelper
}

func TestNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceTestSuite))
}

func init() {
	framework.AuthenticationFlags()
	framework.ControlPlaneFlags()
	framework.TargetClusterFlags()
}

func (s *NamespaceTestSuite) SetupSuite() {
	s.startTime = time.Now()
	s.ctx = context.Background()

	loginCredentials = framework.NewLoginCredentialsFromFlags()

	apiClient, err := api.NewPDSClient(context.Background(), framework.PDSControlPlaneAPI, loginCredentials)
	s.Require().NoError(err, "Could not create Control Plane API client.")

	controlPlane := framework.NewControlPlane(
		s.T(), apiClient,
		controlplane.WithAccountName(framework.PDSAccountName),
		controlplane.WithTenantName(framework.PDSTenantName),
		controlplane.WithProjectName(framework.PDSProjectName),
	)
	s.controlPlane = controlPlane

	framework.InitializePDSHelmChartVersion(s.T(), apiClient)
	s.mustHaveTargetCluster()
	s.crossCluster = crosscluster.NewHelper(s.controlPlane, s.targetCluster, s.startTime)

	targetID := s.controlPlane.MustWaitForDeploymentTarget(context.Background(), s.T(), framework.DeploymentTargetName)
	s.controlPlane.SetTestDeploymentTarget(targetID)
}

func (s *NamespaceTestSuite) TearDownSuite() {
}

func (s *NamespaceTestSuite) mustHaveTargetCluster() {
	token := s.controlPlane.MustGetServiceAccountToken(s.ctx, s.T(), framework.ServiceAccountName)

	tc, err := framework.NewTargetClusterFromFlags(s.controlPlane.TestPDSTenantID, token)
	s.Require().NoError(err, "Cannot create target cluster.")

	s.targetCluster = tc
}
