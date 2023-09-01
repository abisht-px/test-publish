package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/auth"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	certManagerChartVersion = "v1.11.0"
	pdsSystemNamespaceName  = "pds-system"
)

var (
	namePrefix               = fmt.Sprintf("integration-test-%d", time.Now().Unix())
	pdsNamespaceLabelKey     = "pds.portworx.com/available"
	pdsNamespaceLabelValue   = "true"
	auditSecurityLabelKey    = "pod-security.kubernetes.io/audit"
	auditSecurityLabeValue   = "restricted"
	enforceSecurityLabelKey  = "pod-security.kubernetes.io/enforce"
	enforceSecurityLabeValue = "restricted"
	warnSecurityLabelKey     = "pod-security.kubernetes.io/warn"
	warnSecurityLabeValue    = "restricted"
)

type PDSTestSuite struct {
	suite.Suite
	ctx       context.Context
	startTime time.Time

	controlPlane        *controlplane.ControlPlane
	targetCluster       *targetcluster.TargetCluster
	crossCluster        *crosscluster.CrossClusterHelper
	pdsHelmChartVersion string

	config environment
}

func TestPDSSuite(t *testing.T) {
	suite.Run(t, new(PDSTestSuite))
}

func (s *PDSTestSuite) SetupSuite() {
	s.startTime = time.Now()
	s.ctx = context.Background()

	// Try to load .env file from the root of the project.
	err := godotenv.Load("../.env")
	if err == nil {
		s.T().Log("successfully loaded .env file")
	}
	// Perform basic setup with sanity checks.
	env := mustHaveEnvVariables(s.T())
	s.config = env
	s.mustHaveControlPlane(env)
	s.mustHavePDSMetadata(env)
	s.mustHaveTargetCluster(env)
	s.targetCluster.PDSChartConfig.DataServiceTLSEnabled = env.dataServiceTLSEnabled
	s.mustHaveTargetClusterNamespaces(env.pdsNamespaceName)
	s.mustHavePrometheusClient(env)
	if shouldInstallPDSHelmChart(env.pdsHelmChartVersion) {
		s.mustInstallPDSChart(pdsSystemNamespaceName)
		s.mustInstallCertManager()
	}
	targetID := s.controlPlane.MustWaitForDeploymentTarget(s.ctx, s.T(), env.pdsDeploymentTargetName)
	s.controlPlane.SetTestDeploymentTarget(targetID)
	s.controlPlane.MustWaitForTestNamespace(s.ctx, s.T(), env.pdsNamespaceName)

	s.crossCluster = crosscluster.NewHelper(s.controlPlane, s.targetCluster, s.startTime)
}

func (s *PDSTestSuite) TearDownSuite() {
	env := mustHaveEnvVariables(s.T())
	// Do not fail fast on cleanups - we want to clean up as much as possible even on failures.
	s.controlPlane.DeleteTestApplicationTemplates(s.ctx, s.T())
	s.controlPlane.DeleteTestStorageOptions(s.ctx, s.T())
	if shouldInstallPDSHelmChart(env.pdsHelmChartVersion) {
		s.uninstallAgent(env)
		s.controlPlane.DeleteTestDeploymentTarget(s.ctx, s.T())
	}
	wait.For(s.T(), 2*time.Minute, 10*time.Second, func(t tests.T) {
		err := s.targetCluster.DeleteDetachedPXVolumes(s.ctx)
		assert.NoError(t, err, "Cannot delete detached PX volumes.")
	})
}

// mustHavePDSMetadata gets PDS API metadata and stores the PDS helm chart version in the test suite.
func (s *PDSTestSuite) mustHavePDSMetadata(env environment) {
	metadata, resp, err := s.controlPlane.PDS.MetadataApi.ApiMetadataGet(s.ctx).Execute()
	api.RequireNoError(s.T(), resp, err)

	// If user didn't specify the helm chart version, let's use the one configured in PDS API.
	if env.pdsHelmChartVersion == "" || env.pdsHelmChartVersion == "0" {
		s.pdsHelmChartVersion = metadata.GetHelmChartVersion()
	} else {
		s.pdsHelmChartVersion = env.pdsHelmChartVersion
	}
}

func (s *PDSTestSuite) mustCreateTokenSource(env environment) oauth2.TokenSource {
	if env.pdsToken != "" {
		return auth.GetTokenSourceByToken(env.pdsToken)
	}

	tokenSource, err := auth.GetTokenSourceByPassword(
		s.ctx,
		env.controlPlane.LoginCredentials.TokenIssuerURL,
		env.controlPlane.LoginCredentials.IssuerClientID,
		env.controlPlane.LoginCredentials.IssuerClientSecret,
		env.controlPlane.LoginCredentials.Username,
		env.controlPlane.LoginCredentials.Password,
	)
	s.Require().NoError(err, "Cannot create token source.")
	return tokenSource
}

func (s *PDSTestSuite) mustHaveControlPlane(env environment) {
	apiClient, err := api.NewPDSClient(context.Background(), env.controlPlane.ControlPlaneAPI, env.controlPlane.LoginCredentials)
	s.Require().NoError(err, "Could not create Control Plane API client.")

	controlPlane := controlplane.New(apiClient)
	controlPlane.MustInitializeTestData(s.ctx, s.T(),
		env.controlPlane.AccountName,
		env.controlPlane.TenantName,
		env.controlPlane.ProjectName,
		namePrefix)
	s.controlPlane = controlPlane

	s.mustHavePrometheusClient(env)
}

func (s *PDSTestSuite) mustHavePrometheusClient(env environment) {
	tokenSource := s.mustCreateTokenSource(env)
	s.controlPlane.MustSetupPrometheus(s.T(), env.controlPlane.PrometheusAPI, tokenSource)
}

func (s *PDSTestSuite) mustHaveTargetCluster(env environment) {
	token := s.controlPlane.MustGetServiceAccountToken(s.ctx, s.T(), env.pdsServiceAccountName)
	pdsChartConfig := targetcluster.PDSChartConfig{
		Version:              s.pdsHelmChartVersion,
		TenantID:             s.controlPlane.TestPDSTenantID,
		Token:                token,
		ControlPlaneAPI:      env.controlPlane.ControlPlaneAPI,
		DeploymentTargetName: env.pdsDeploymentTargetName,
	}
	certManagerChartConfig := targetcluster.CertManagerChartConfig{
		Version: certManagerChartVersion,
	}
	tc, err := targetcluster.NewTargetCluster(s.ctx, env.targetKubeconfig, pdsChartConfig, certManagerChartConfig)
	s.Require().NoError(err, "Cannot create target cluster.")
	s.targetCluster = tc
}

func (s *PDSTestSuite) mustInstallPDSChart(namespaceName string) {
	labels := map[string]string{
		pdsNamespaceLabelKey:    pdsNamespaceLabelValue,
		auditSecurityLabelKey:   auditSecurityLabeValue,
		enforceSecurityLabelKey: enforceSecurityLabeValue,
		warnSecurityLabelKey:    warnSecurityLabeValue,
	}
	namespace, err := s.targetCluster.GetNamespace(s.ctx, namespaceName)
	if err == nil {
		namespace.Labels = labels
		_, err = s.targetCluster.UpdateNamespace(s.ctx, namespace)
		s.Require().NoError(err, "Updating namespace %s", namespace.Name)

	} else if k8serrors.IsNotFound(err) {
		k8sns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   namespaceName,
				Labels: labels,
			},
		}
		_, err = s.targetCluster.CreateNamespace(s.ctx, k8sns)
		s.Require().NoError(err, "Creating namespace %s", k8sns.Name)
	} else {
		s.Require().NoError(err)
	}
	err = s.targetCluster.InstallPDSChart(s.ctx)
	s.Require().NoError(err, "Failed to install PDS helm chart")
}

func (s *PDSTestSuite) mustInstallCertManager() {
	err := s.targetCluster.InstallCertManagerChart(s.ctx)
	s.Require().NoError(err, "Failed to install CertManager helm chart")
}

func (s *PDSTestSuite) uninstallAgent(env environment) {
	err := s.targetCluster.UninstallCertManagerChart(s.ctx)
	s.NoError(err, "Cannot uninstall cert manager.")
	err = s.targetCluster.DeleteCRDs(s.ctx)
	s.NoError(err, "Cannot delete CRDs.")
	err = s.targetCluster.UninstallPDSChart(s.ctx)
	s.NoError(err, "Cannot uninstall agent.")
	err = s.targetCluster.DeleteClusterRoles(s.ctx)
	s.NoError(err, "Cannot delete cluster roles.")
	err = s.targetCluster.DeletePVCs(s.ctx, env.pdsNamespaceName)
	s.NoError(err, "Cannot delete PVCs.")
	err = s.targetCluster.DeleteStorageClasses(s.ctx)
	s.NoError(err, "Cannot delete storage classes.")
	err = s.targetCluster.DeleteReleasedPVs(s.ctx)
	s.NoError(err, "Cannot delete released PVs.")
	wait.For(s.T(), 5*time.Minute, 10*time.Second, func(t tests.T) {
		err := s.targetCluster.DeleteDetachedPXVolumes(s.ctx)
		assert.NoError(t, err, "Cannot delete detached PX volumes.")
	})
	err = s.targetCluster.DeletePXCloudCredentials(s.ctx)
	s.NoError(err, "Cannot delete PX cloud credentials.")
}

func (s *PDSTestSuite) mustHaveTargetClusterNamespaces(name string) {
	namespace, err := s.targetCluster.GetNamespace(s.ctx, name)
	if err != nil {
		k8sns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					pdsNamespaceLabelKey: pdsNamespaceLabelValue,
				},
			},
		}
		namespace, err = s.targetCluster.CreateNamespace(s.ctx, k8sns)
		s.Require().NoError(err, "Creating namespace %s", k8sns.Name)
	}
	labelValue, ok := namespace.Labels[pdsNamespaceLabelKey]
	if !ok || labelValue != pdsNamespaceLabelValue {
		namespace.Labels = map[string]string{
			pdsNamespaceLabelKey: pdsNamespaceLabelValue,
		}
		_, err = s.targetCluster.UpdateNamespace(s.ctx, namespace)
		s.Require().NoError(err, "Updating namespace %s", namespace.Name)
	}
}
