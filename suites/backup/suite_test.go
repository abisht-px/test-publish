package backup_test

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/suites/framework"
	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
)

var (
	ctx              context.Context
	backupTargetCfg  framework.BackupTargetConfig
	controlPlane     *controlplane.ControlPlane
	targetCluster    *targetcluster.TargetCluster
	crossCluster     *crosscluster.CrossClusterHelper
	cleanupNamespace bool
)

type BackupTestSuite struct {
	suite.Suite
}

func init() {
	framework.AuthenticationFlags()
	framework.ControlPlaneFlags()
	framework.TargetClusterFlags()
	framework.BackupCredentialFlags()
}

func TestBackupTestSuite(t *testing.T) {
	suite.Run(t, new(BackupTestSuite))
}

func (s *BackupTestSuite) SetupSuite() {
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
		framework.TestNamespace = framework.NewRandomName("ns-backup")
		framework.EnsureTestNamespace(s.T(), targetCluster, framework.TestNamespace)
		cleanupNamespace = true
	}

	cp.MustWaitForTestNamespace(context.Background(), s.T(), framework.TestNamespace)

	crossCluster = crosscluster.NewHelper(controlPlane, targetCluster, time.Now())
}

func (s *BackupTestSuite) TearDownSuite() {
	if cleanupNamespace {
		framework.CleanupTestNamespace(s.T(), targetCluster, framework.TestNamespace)
	}

	controlPlane.DeleteTestApplicationTemplates(context.Background(), s.T())
	controlPlane.DeleteTestStorageOptions(context.Background(), s.T())

}

func deleteBackupWithWorkaround(t *testing.T, backup *pds.ModelsBackup, namespace string) {
	// TODO(DS-5732): Once bug https://portworx.atlassian.net/browse/DS-5732 is fixed then call MustDeleteBackup
	// 		with localOnly=false and remove the additional call of DeletePDSBackup.
	controlPlane.MustDeleteBackup(ctx, t, backup.GetId(), true)

	err := targetCluster.DeletePDSBackup(ctx, namespace, backup.GetClusterResourceName())
	require.NoError(t, err)
}

func getBackupJobID(backupJob *backupsv1.BackupJob) (string, error) {
	backupJobID := string(backupJob.GetUID())
	if backupJobID == "" {
		return "", errors.New("backupJob id is empty")
	}

	return backupJobID, nil
}
