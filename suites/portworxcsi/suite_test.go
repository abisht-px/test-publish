package portworxcsi_test

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

type PortworxCSITestSuite struct {
	ctx context.Context
	//backupTargetCfg  framework.BackupTargetConfig
	controlPlane     *controlplane.ControlPlane
	targetCluster    *targetcluster.TargetCluster
	crossCluster     *crosscluster.CrossClusterHelper
	cleanupNamespace bool
	suite.Suite
}

func init() {
	framework.AuthenticationFlags()
	framework.ControlPlaneFlags()
	framework.TargetClusterFlags()
	//framework.BackupCredentialFlags()
}

func TestPortworxCSITestSuite(t *testing.T) {
	suite.Run(t, new(PortworxCSITestSuite))
}

func (s *PortworxCSITestSuite) SetupSuite() {
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
	)
	s.controlPlane = cp

	//backupTargetCfg = framework.NewBackupTargetConfigFromFlags()

	token := cp.MustGetServiceAccountToken(context.Background(), s.T(), framework.ServiceAccountName)
	framework.InitializePDSHelmChartVersion(s.T(), apiClient)

	s.targetCluster, err = framework.NewTargetClusterFromFlags(cp.TestPDSTenantID, token)
	require.NoError(s.T(), err, "Cannot create target cluster.")

	targetID := cp.MustWaitForDeploymentTarget(context.Background(), s.T(), framework.DeploymentTargetName)
	cp.SetTestDeploymentTarget(targetID)

	if framework.TestNamespace == "" {
		framework.TestNamespace = framework.NewRandomName("ns-backup")
		framework.EnsureTestNamespace(s.T(), s.targetCluster, framework.TestNamespace)
		s.cleanupNamespace = true
	}

	cp.MustWaitForTestNamespace(context.Background(), s.T(), framework.TestNamespace)

	s.crossCluster = crosscluster.NewHelper(s.controlPlane, s.targetCluster, time.Now())
}

func (s *PortworxCSITestSuite) TearDownSuite() {
	if s.cleanupNamespace {
		framework.CleanupTestNamespace(s.T(), s.targetCluster, framework.TestNamespace)
	}

	s.controlPlane.DeleteTestApplicationTemplates(context.Background(), s.T())
	s.controlPlane.DeleteTestStorageOptions(context.Background(), s.T())

}
