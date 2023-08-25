package capabilities_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/suites/framework"
)

type CapabilitiesTestSuite struct {
	ctx           context.Context
	controlPlane  *controlplane.ControlPlane
	targetCluster *targetcluster.TargetCluster
	crossCluster  *crosscluster.CrossClusterHelper
	suite.Suite
}

func init() {
	framework.AuthenticationFlags()
	framework.ControlPlaneFlags()
	framework.TargetClusterFlags()
}

func TestCapabilitiesTestSuite(t *testing.T) {
	suite.Run(t, new(CapabilitiesTestSuite))
}

func (s *CapabilitiesTestSuite) SetupSuite() {
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
		controlplane.WithLoadImageVersions(),
		controlplane.WithCreateTemplatesAndStorageOptions(framework.NewRandomName("temp")),
		controlplane.WithPrometheusClient(framework.PDSControlPlaneAPI, framework.NewLoginCredentialsFromFlags()),
	)
	s.controlPlane = cp

	token := cp.MustGetServiceAccountToken(context.Background(), s.T(), framework.ServiceAccountName)
	framework.InitializePDSHelmChartVersion(s.T(), apiClient)

	s.targetCluster, err = framework.NewTargetClusterFromFlags(cp.TestPDSTenantID, token)
	require.NoError(s.T(), err, "Cannot create target cluster.")

	targetID := cp.MustWaitForDeploymentTarget(context.Background(), s.T(), framework.DeploymentTargetName)
	cp.SetTestDeploymentTarget(targetID)

	s.crossCluster = crosscluster.NewHelper(s.controlPlane, s.targetCluster, time.Now())

}

func (s *CapabilitiesTestSuite) TearDownSuite() {
	s.controlPlane.DeleteTestApplicationTemplates(context.Background(), s.T())
	s.controlPlane.DeleteTestStorageOptions(context.Background(), s.T())
}
