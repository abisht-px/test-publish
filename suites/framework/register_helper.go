package framework

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
)

const (
	DefaultPollPeriod               = 10 * time.Second
	DefaultTimeout                  = 5 * time.Minute
	ErrStrHelmReleaseAlreadyInUse   = "cannot re-use a name that is still in use"
	ErrStrHelmReleaseAlreadyDeleted = "is already deleted"
	ErrStrHelmReleaseNotFound       = "release: not found"
)

var pdsNamespaceLabel = map[string]string{
	"pds.portworx.com/available": "true",
}

func EnsurePDSNamespace(
	ctx context.Context,
	tc *targetcluster.TargetCluster,
) error {
	ns, err := tc.GetNamespace(ctx, DefaultPDSNamespace)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			_, err = tc.CreateNamespace(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   DefaultPDSNamespace,
					Labels: pdsNamespaceLabel,
				},
			})
			if err != nil {
				return errors.Wrap(err, "create namespace")
			}

			return nil
		}

		return errors.Wrap(err, "get namespace")
	}

	ns.SetLabels(pdsNamespaceLabel)

	if _, err := tc.UpdateNamespace(ctx, ns); err != nil {
		return errors.Wrap(err, "update namespace")
	}

	return nil
}

func CleanupTargetCluster(
	ctx context.Context,
	t *testing.T,
	tc *targetcluster.TargetCluster,
) {
	t.Run("uninstall agents", func(t *testing.T) {
		assert.NoError(t, UninstallPDSAgents(ctx, tc), "uninstall agents")
	})

	t.Run("cleanup PDS Namespace", func(t *testing.T) {
		assert.NoError(t, CleanupNamespace(ctx, tc, DefaultPDSNamespace), "uninstall agents")
	})
}

func CleanupNamespace(
	ctx context.Context,
	tc *targetcluster.TargetCluster,
	namespace string,
) error {
	err := tc.DeleteNamespace(ctx, namespace)
	if err != nil && k8serrors.IsNotFound(err) {
		return nil
	}

	return errors.Wrapf(err, "delete namespace %s", namespace)
}

func UninstallPDSAgents(
	ctx context.Context,
	tc *targetcluster.TargetCluster,
) error {
	err := tc.DeleteCRDs(ctx)
	if err != nil {
		return errors.Wrap(err, "delete CRDs")
	}

	err = tc.UninstallPDSChart(ctx)
	if err != nil && !ErrorContainsAnyOfMsg(err, ErrStrHelmReleaseAlreadyDeleted, ErrStrHelmReleaseNotFound) {
		return errors.Wrap(err, "uninstall PDS chart")
	}

	err = tc.DeleteClusterRoles(ctx)
	if err != nil {
		return errors.Wrap(err, "delete cluster roles")
	}

	err = tc.DeletePVCs(ctx, DefaultPDSNamespace)
	if err != nil {
		return errors.Wrap(err, "delete PVCs")
	}

	err = tc.DeleteStorageClasses(ctx)
	if err != nil {
		return errors.Wrap(err, "delete storage classes")
	}

	err = tc.DeleteReleasedPVs(ctx)
	if err != nil {
		return errors.Wrap(err, "delete released PVs")
	}

	err = tc.DeletePXCloudCredentials(ctx)
	if err != nil {
		return errors.Wrap(err, "delete PX cloud credentials")
	}

	return nil
}

func EnsureNamespaceCleanup(
	ctx context.Context,
	s *suite.Suite,
	tc *targetcluster.TargetCluster,
	namespace string,
) {
	s.Eventuallyf(func() bool {
		return CleanupNamespace(ctx, tc, namespace) == nil
	},
		DefaultTimeout,
		DefaultPollPeriod,
		"Namespace %s still exists", namespace,
	)
}

func InstallCertManager(
	ctx context.Context,
	tc *targetcluster.TargetCluster,
) error {
	err := tc.InstallCertManagerChart(ctx)
	if err != nil && !ErrorContainsAnyOfMsg(err, ErrStrHelmReleaseAlreadyInUse) {
		return errors.Wrap(err, "Failed to install CertManager helm chart")
	}

	return nil
}

func UninstallCertManager(
	ctx context.Context,
	tc *targetcluster.TargetCluster,
) error {
	if err := tc.UninstallCertManagerChart(ctx); err != nil &&
		!ErrorContainsAnyOfMsg(err, ErrStrHelmReleaseAlreadyDeleted, ErrStrHelmReleaseNotFound) {
		return errors.Wrap(err, "Failed to uninstall CertManager helm chart")
	}

	if err := CleanupNamespace(ctx, tc, DefaultCertManagerNamespace); err != nil {
		return errors.Wrap(err, "Failed to delete cert manager namespace")
	}

	return nil
}

func RegisterTargetCluster(s *suite.Suite, cp *controlplane.ControlPlane, tc *targetcluster.TargetCluster) {
	s.Run("Install Cert Manager Chart", func() {
		require.NoError(s.T(), InstallCertManager(context.Background(), tc))
	})

	s.Run("Ensure PDS Namespace", func() {
		require.NoError(s.T(), EnsurePDSNamespace(context.Background(), tc))
	})

	s.Run(fmt.Sprintf("Install PDS Chart v%s", PDSHelmChartVersion), func() {
		require.NoError(s.T(), tc.InstallPDSChart(context.Background()))
	})

	s.Run(fmt.Sprintf("Verify deployment target %s", DeploymentTargetName), func() {
		targetID := cp.MustWaitForDeploymentTarget(context.Background(), s.T(), DeploymentTargetName)
		cp.SetTestDeploymentTarget(targetID)
		cp.MustWaitForTestNamespace(context.Background(), s.T(), DefaultPDSNamespace)
	})
}

func DeregisterTargetCluster(s *suite.Suite, cp *controlplane.ControlPlane, tc *targetcluster.TargetCluster) {
	targetID := cp.GetDeploymentTargetID(context.Background(), s.T(), DeploymentTargetName)
	cp.SetTestDeploymentTarget(targetID)

	s.Run("Cleanup deployments from CP", func() {
		CleanupDeploymentsForTheCluster(s.T(), cp)
	})

	s.Run("Cleanup Target Cluster", func() {
		CleanupTargetCluster(context.Background(), s.T(), tc)
	})

	s.Run("Uninstall Cert Manager Chart", func() {
		assert.NoError(s.T(), UninstallCertManager(context.Background(), tc))
	})

	s.Run(fmt.Sprintf("Delete target cluster %s from CP", DeploymentTargetName), func() {
		cp.DeleteTestDeploymentTarget(context.Background(), s.T())
	})

	s.Run("Ensure PDS Namespace Cleanup", func() {
		EnsureNamespaceCleanup(context.Background(), s, tc, DefaultPDSNamespace)
	})

	s.Run("Ensure Cert Manager Namespace Cleanup", func() {
		EnsureNamespaceCleanup(context.Background(), s, tc, DefaultCertManagerNamespace)
	})
}

func EnsureTestNamespace(
	t *testing.T,
	tc *targetcluster.TargetCluster,
	namespaceName string,
) {
	_, err := tc.CreateNamespace(context.Background(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespaceName,
			Labels: pdsNamespaceLabel,
		},
	})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		require.NoErrorf(t, err, "create test namespace %s", namespaceName)
	}
}

func CleanupTestNamespace(
	t *testing.T,
	tc *targetcluster.TargetCluster,
	namespaceName string,
) {
	err := tc.DeleteNamespace(context.Background(), namespaceName)
	if err != nil && !k8serrors.IsNotFound(err) {
		require.NoErrorf(t, err, "delete test namespace %s", namespaceName)
	}
}

func CleanupDeploymentsForTheCluster(
	t *testing.T,
	cp *controlplane.ControlPlane,
) {
	ctx := context.Background()

	backupJobIDList, resp, err := cp.ListBackupJobsInProject(
		ctx, cp.TestPDSProjectID,
		controlplane.WithListBackupJobsInDeploymentTarget(cp.DeploymentTargetID()),
	)
	assert.NoError(t, api.ExtractErrorDetails(resp, err))

	for _, each := range backupJobIDList {
		cp.MustDeleteBackupJobByID(ctx, t, *each.Id)
	}

	deploymentList, err := cp.ListDeploymentsForDeploymentTarget(ctx, cp.TestPDSProjectID, cp.DeploymentTargetID())
	assert.NoError(t, err)

	for _, each := range deploymentList {
		backupList, err := cp.ListBackupsByDeploymentID(ctx, *each.Id)
		assert.NoError(t, err)

		for _, each := range backupList {
			cp.MustDeleteBackup(ctx, t, *each.Id, true)
		}

		cp.MustRemoveDeployment(ctx, t, *each.Id)
	}
}
