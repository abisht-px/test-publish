package deployment_test

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

var (
	ctx              context.Context
	backupTargetCfg  framework.BackupTargetConfig
	controlPlane     *controlplane.ControlPlane
	targetCluster    *targetcluster.TargetCluster
	crossCluster     *crosscluster.CrossClusterHelper
	cleanupNamespace bool
)

type DeploymentTestSuite struct {
	suite.Suite
}

func init() {
	framework.AuthenticationFlags()
	framework.ControlPlaneFlags()
	framework.TargetClusterFlags()
	framework.BackupCredentialFlags()
}

func TestDeploymentTestSuite(t *testing.T) {
	suite.Run(t, new(DeploymentTestSuite))
}

func (s *DeploymentTestSuite) SetupSuite() {
	ctx = context.Background()

	apiClient, err := api.NewPDSClient(
		ctx,
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
	)
	controlPlane = cp

	backupTargetCfg = framework.NewBackupTargetConfigFromFlags()

	token := cp.MustGetServiceAccountToken(context.Background(), s.T(), framework.ServiceAccountName)
	framework.InitializePDSHelmChartVersion(s.T(), apiClient)

	targetCluster, err = framework.NewTargetClusterFromFlags(cp.TestPDSTenantID, token)
	require.NoError(s.T(), err, "Cannot create target cluster.")

	targetID := cp.MustWaitForDeploymentTarget(context.Background(), s.T(), framework.DeploymentTargetName)
	cp.SetTestDeploymentTarget(targetID)

	if framework.TestNamespace == "" {
		framework.TestNamespace = framework.NewRandomName("ns-d")
		framework.EnsureTestNamespace(s.T(), targetCluster, framework.TestNamespace)
		cleanupNamespace = true
	}

	cp.MustWaitForTestNamespace(context.Background(), s.T(), framework.TestNamespace)

	crossCluster = crosscluster.NewHelper(controlPlane, targetCluster, time.Now())
}

func (s *DeploymentTestSuite) TearDownSuite() {
	if cleanupNamespace {
		framework.CleanupTestNamespace(s.T(), targetCluster, framework.TestNamespace)
	}

	controlPlane.DeleteTestApplicationTemplates(context.Background(), s.T())
	controlPlane.DeleteTestStorageOptions(context.Background(), s.T())

}
