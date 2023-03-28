package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/auth"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/helminstaller"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

var (
	namePrefix             = fmt.Sprintf("integration-test-%d", time.Now().Unix())
	pdsNamespaceLabelKey   = "pds.portworx.com/available"
	pdsNamespaceLabelValue = "true"
)

type PDSTestSuite struct {
	suite.Suite
	ctx       context.Context
	startTime time.Time

	controlPlane        *controlplane.ControlPlane
	targetCluster       *targetcluster.TargetCluster
	crossCluster        *crosscluster.CrossClusterHelper
	pdsAgentInstallable *helminstaller.InstallableHelm
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
	s.mustHaveTargetCluster(env)
	s.mustHaveTargetClusterNamespaces(env.pdsNamespaceName)
	s.mustHaveControlPlane(env)
	s.mustHavePDSMetadata(env)
	s.mustHavePrometheusClient(env)
	if shouldInstallPDSHelmChart(s.pdsHelmChartVersion) {
		s.mustInstallAgent(env)
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
}

// mustHavePDSMetadata gets PDS API metadata and stores the PDS helm chart version in the test suite.
func (s *PDSTestSuite) mustHavePDSMetadata(env environment) {
	metadata, resp, err := s.controlPlane.PDS.MetadataApi.ApiMetadataGet(s.ctx).Execute()
	api.RequireNoError(s.T(), resp, err)

	// If user didn't specify the helm chart version, let's use the one configured in PDS API.
	if env.pdsHelmChartVersion == "" {
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
	tc, err := targetcluster.NewTargetCluster(s.ctx, env.targetKubeconfig)
	s.Require().NoError(err, "Cannot create target cluster.")
	s.targetCluster = tc
}

func (s *PDSTestSuite) mustInstallAgent(env environment) {
	token := s.controlPlane.MustGetServiceAccountToken(s.ctx, s.T(), env.pdsServiceAccountName)

	provider, err := helminstaller.NewHelmProviderPDS()
	s.Require().NoError(err, "Cannot create agent installer provider.")

	pdsChartConfig := helminstaller.NewPDSChartConfig(s.pdsHelmChartVersion, s.controlPlane.TestPDSTenantID, token, env.controlPlane.ControlPlaneAPI, env.pdsDeploymentTargetName)

	installer, err := provider.Installer(env.targetKubeconfig, pdsChartConfig)
	s.Require().NoError(err, "Cannot get agent installer for version constraints %s.", pdsChartConfig.VersionConstraints)

	err = installer.Install(s.ctx)
	s.Require().NoError(err, "Cannot install agent for version %s.", installer.Version())
	s.pdsAgentInstallable = installer
}

func (s *PDSTestSuite) uninstallAgent(env environment) {
	err := s.targetCluster.DeleteCRDs(s.ctx)
	s.NoError(err, "Cannot delete CRDs.")
	err = s.pdsAgentInstallable.Uninstall(s.ctx)
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

func (s *PDSTestSuite) mustVerifyMetrics(t *testing.T, deploymentID string) {
	deployment, resp, err := s.controlPlane.PDS.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := s.controlPlane.PDS.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)
	dataServiceType := dataService.GetName()

	require.Contains(t, dataServiceExpectedMetrics, dataServiceType, "%s data service has no defined expected metrics")
	expectedMetrics := dataServiceExpectedMetrics[dataServiceType]

	var missingMetrics []model.LabelValue
	for _, expectedMetric := range expectedMetrics {
		// Add deployment ID to the metric label filter.
		pdsDeploymentIDMatch := parser.MustLabelMatcher(labels.MatchEqual, "pds_deployment_id", deploymentID)
		expectedMetric.LabelMatchers = append(expectedMetric.LabelMatchers, pdsDeploymentIDMatch)

		queryResult, _, err := s.controlPlane.Prometheus.Query(s.ctx, expectedMetric.String(), time.Now())
		require.NoError(t, err, "prometheus: query error")

		require.IsType(t, model.Vector{}, queryResult, "prometheus: wrong result model")
		queryResultMetrics := queryResult.(model.Vector)

		if len(queryResultMetrics) == 0 {
			missingMetrics = append(missingMetrics, model.LabelValue(expectedMetric.Name))
		}
	}

	require.Equalf(t, len(missingMetrics), 0, "%s: prometheus missing %d/%d metrics: %v", dataServiceType, len(missingMetrics), len(expectedMetrics), missingMetrics)
}
