package targetcluster_test

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

var dsVersions framework.DSVersionMatrix

type TargetClusterTestSuite struct {
	suite.Suite

	ctx              context.Context
	startTime        time.Time
	controlPlane     *controlplane.ControlPlane
	targetCluster    *targetcluster.TargetCluster
	crossCluster     *crosscluster.CrossClusterHelper
	cleanupNamespace bool
}

func init() {
	framework.AuthenticationFlags()
	framework.ControlPlaneFlags()
	framework.TargetClusterFlags()
	framework.DataserviceFlags()
}

func TestTargetClusterTestSuite(t *testing.T) {
	suite.Run(t, new(TargetClusterTestSuite))
}

func (s *TargetClusterTestSuite) SetupSuite() {
	s.startTime = time.Now()
	s.ctx = context.Background()

	dsVersionMatrix, err := framework.NewDSVersionMatrixFromFlags()
	s.Require().NoError(err, "load dataservice versions")
	dsVersions = dsVersionMatrix

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

	if framework.TestNamespace == "" {
		framework.TestNamespace = framework.NewRandomName("ns-tc")
		framework.EnsureTestNamespace(s.T(), s.targetCluster, framework.TestNamespace)
		s.cleanupNamespace = true
	}

	cp.MustWaitForTestNamespace(context.Background(), s.T(), framework.TestNamespace)

	s.crossCluster = crosscluster.NewHelper(s.controlPlane, s.targetCluster, s.startTime)
}

func (s *TargetClusterTestSuite) TearDownSuite() {
	if s.cleanupNamespace {
		framework.CleanupTestNamespace(s.T(), s.targetCluster, framework.TestNamespace)
	}

	s.controlPlane.DeleteTestApplicationTemplates(context.Background(), s.T())
	s.controlPlane.DeleteTestStorageOptions(context.Background(), s.T())
}
