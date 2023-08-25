package dataservices_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/suites/framework"

	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
)

func init() {
	framework.AuthenticationFlags()
	framework.ControlPlaneFlags()
	framework.TargetClusterFlags()
	framework.BackupCredentialFlags()
	framework.DataserviceFlags()
}

func TestBackupRestoreSuite(t *testing.T) {
	suite.Run(t, new(BackupRestoreSuite))
}

func TestDataservicesSuite(t *testing.T) {
	suite.Run(t, new(Dataservices))
}

func TestScaleSuite(t *testing.T) {
	suite.Run(t, new(ScaleSuite))
}

func TestMetricsSuite(t *testing.T) {
	suite.Run(t, new(MetricsSuite))
}

func SetupSuite(t *testing.T, prefix string, options ...controlplane.InitializeOption) (
	*controlplane.ControlPlane,
	*targetcluster.TargetCluster,
	*crosscluster.CrossClusterHelper,
) {
	ctx := context.Background()

	apiClient, err := api.NewPDSClient(
		ctx,
		framework.PDSControlPlaneAPI,
		framework.NewLoginCredentialsFromFlags(),
	)
	require.NoError(t, err, "could not create Control Plane API client")

	controlPlane := framework.NewControlPlane(
		t,
		apiClient,
		options...,
	)

	token := controlPlane.MustGetServiceAccountToken(ctx, t, framework.ServiceAccountName)
	framework.InitializePDSHelmChartVersion(t, apiClient)

	targetCluster, err := framework.NewTargetClusterFromFlags(controlPlane.TestPDSTenantID, token)
	require.NoError(t, err, "Cannot create target cluster.")

	targetID := controlPlane.MustWaitForDeploymentTarget(ctx, t, framework.DeploymentTargetName)
	controlPlane.SetTestDeploymentTarget(targetID)

	if framework.TestNamespace == "" {
		framework.TestNamespace = framework.NewRandomName(prefix)
		framework.EnsureTestNamespace(t, targetCluster, framework.TestNamespace)
	}

	controlPlane.MustWaitForTestNamespace(ctx, t, framework.TestNamespace)

	crossCluster := crosscluster.NewHelper(controlPlane, targetCluster, time.Now())

	return controlPlane, targetCluster, crossCluster
}

func TearDownSuite(t *testing.T, cp *controlplane.ControlPlane, tc *targetcluster.TargetCluster) {
	cp.DeleteTestApplicationTemplates(context.Background(), t)
	cp.DeleteTestStorageOptions(context.Background(), t)

	// wait.For(t, 5*time.Minute, 10*time.Second, func(t tests.T) {
	// 	err := tc.DeleteDetachedPXVolumes(context.Background())
	// 	assert.NoError(t, err, "Cannot delete detached PX volumes.")
	// })
}
