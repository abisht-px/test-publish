package tls_test

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
	controlPlane     *controlplane.ControlPlane
	targetCluster    *targetcluster.TargetCluster
	crossCluster     *crosscluster.CrossClusterHelper
	dsVersions       framework.DSVersionMatrix
	cleanupNamespace bool
)

type TLSSuite struct {
	suite.Suite
}

func init() {
	framework.AuthenticationFlags()
	framework.ControlPlaneFlags()
	framework.TargetClusterFlags()
	framework.BackupCredentialFlags()
	framework.DataserviceFlags()
}

func TestTLSSuite(t *testing.T) {
	suite.Run(t, new(TLSSuite))
}

func (s *TLSSuite) SetupSuite() {
	ctx = context.Background()

	dsVersionMatrix, err := framework.NewDSVersionMatrixFromFlags()
	s.Require().NoError(err, "load dataservice versions")
	dsVersions = dsVersionMatrix

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

	token := cp.MustGetServiceAccountToken(context.Background(), s.T(), framework.ServiceAccountName)
	framework.InitializePDSHelmChartVersion(s.T(), apiClient)

	targetCluster, err = framework.NewTargetClusterFromFlags(cp.TestPDSTenantID, token)
	require.NoError(s.T(), err, "Cannot create target cluster.")

	targetID := cp.MustWaitForDeploymentTarget(context.Background(), s.T(), framework.DeploymentTargetName)
	cp.SetTestDeploymentTarget(targetID)

	if framework.TestNamespace == "" {
		framework.TestNamespace = framework.NewRandomName("ns-tls")
		framework.EnsureTestNamespace(s.T(), targetCluster, framework.TestNamespace)
		cleanupNamespace = true
	}

	cp.MustWaitForTestNamespace(context.Background(), s.T(), framework.TestNamespace)

	crossCluster = crosscluster.NewHelper(controlPlane, targetCluster, time.Now())
}

func (s *TLSSuite) TearDownSuite() {
	if cleanupNamespace {
		framework.CleanupTestNamespace(s.T(), targetCluster, framework.TestNamespace)
	}

	controlPlane.DeleteTestApplicationTemplates(context.Background(), s.T())
	controlPlane.DeleteTestStorageOptions(context.Background(), s.T())

}
