package test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/auth"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/crosscluster"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/helminstaller"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/internal/prometheus"
	"github.com/portworx/pds-integration-test/internal/random"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	waiterShortRetryInterval                     = time.Second * 1
	waiterRetryInterval                          = time.Second * 10
	waiterDeploymentTargetNameExistsTimeout      = time.Second * 90
	waiterNamespaceExistsTimeout                 = time.Second * 30
	waiterDeploymentTargetStatusHealthyTimeout   = time.Minute * 10
	waiterDeploymentTargetStatusUnhealthyTimeout = time.Second * 300
	waiterDeploymentStatusHealthyTimeout         = time.Second * 600
	waiterLoadBalancerServicesReady              = time.Second * 300
	waiterStatefulSetReadyAndUpdatedReplicas     = time.Second * 600
	waiterBackupStatusSucceededTimeout           = time.Second * 300
	waiterBackupTargetSyncedTimeout              = time.Second * 60
	waiterDeploymentStatusRemovedTimeout         = time.Second * 300
	waiterLoadTestJobFinishedTimeout             = time.Second * 300
	waiterHostCheckFinishedTimeout               = time.Second * 60
	waiterAllHostsAvailableTimeout               = time.Second * 600
	waiterCoreDNSRestartedTimeout                = time.Second * 30

	pdsAPITimeFormat = "2006-01-02T15:04:05.999999Z"
)

var (
	namePrefix                 = fmt.Sprintf("integration-test-%d", time.Now().Unix())
	pdsUserInRedisIntroducedAt = time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC)
	pdsNamespaceLabelKey       = "pds.portworx.com/available"
	pdsNamespaceLabelValue     = "true"
)

type PDSTestSuite struct {
	suite.Suite
	ctx       context.Context
	startTime time.Time

	controlPlane        *controlplane.ControlPlane
	targetCluster       *targetcluster.TargetCluster
	crossCluster        *crosscluster.CrossClusterHelper
	prometheusClient    prometheusv1.API
	pdsAgentInstallable *helminstaller.InstallableHelmPDS
	pdsHelmChartVersion string

	testPDSAgentToken string
	config            environment
	tokenSource       oauth2.TokenSource
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
	s.mustHaveAuthorization(env)
	s.mustHaveControlPlane(env)
	s.mustHavePDSMetadata(env)
	s.mustHavePrometheusClient(env)
	if shouldInstallPDSHelmChart(s.pdsHelmChartVersion) {
		s.mustHavePDStestServiceAccount(env)
		s.mustHavePDStestAgentToken(env)
		s.mustInstallAgent(env)
	}
	s.mustWaitForPDSTestDeploymentTarget(env)
	s.controlPlane.MustWaitForTestNamespace(s.ctx, s.T(), env.pdsNamespaceName)

	s.crossCluster = crosscluster.NewHelper(s.controlPlane, s.targetCluster)
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
	metadata, resp, err := s.controlPlane.API.MetadataApi.ApiMetadataGet(s.ctx).Execute()
	api.RequireNoError(s.T(), resp, err)

	// If user didn't specify the helm chart version, let's use the one configured in PDS API.
	if env.pdsHelmChartVersion == "" {
		s.pdsHelmChartVersion = metadata.GetHelmChartVersion()
	} else {
		s.pdsHelmChartVersion = env.pdsHelmChartVersion
	}
}

func (s *PDSTestSuite) mustWaitForPDSTestDeploymentTarget(env environment) {
	wait.For(s.T(), waiterDeploymentTargetNameExistsTimeout, waiterRetryInterval, func(t tests.T) {
		var err error
		s.controlPlane.TestPDSDeploymentTargetID, err = s.controlPlane.API.GetDeploymentTargetIDByName(s.ctx, s.controlPlane.TestPDSTenantID, env.pdsDeploymentTargetName)
		require.NoErrorf(t, err, "PDS deployment target %q does not exist.", env.pdsDeploymentTargetName)
	})

	wait.For(s.T(), waiterDeploymentTargetStatusHealthyTimeout, waiterRetryInterval, func(t tests.T) {
		err := s.controlPlane.API.CheckDeploymentTargetHealth(s.ctx, s.controlPlane.TestPDSDeploymentTargetID)
		require.NoErrorf(t, err, "Deployment target %q is not healthy.", s.controlPlane.TestPDSDeploymentTargetID)
	})
}

// mustHavePDStestServiceAccount finds PDS Service account in Test PDS tenant with name set in environment and stores its ID as "Test PDS Service Account".
func (s *PDSTestSuite) mustHavePDStestServiceAccount(env environment) {
	// TODO: Use service account name query filters
	serviceAccounts, resp, err := s.controlPlane.API.ServiceAccountsApi.ApiTenantsIdServiceAccountsGet(s.ctx, s.controlPlane.TestPDSTenantID).Execute()
	api.RequireNoError(s.T(), resp, err)
	s.Require().NotEmpty(serviceAccounts, "PDS API must return at least one tenant.")

	var testPDSServiceAccountID string
	for _, serviceAccount := range serviceAccounts.GetData() {
		if serviceAccount.GetName() == env.pdsServiceAccountName {
			testPDSServiceAccountID = serviceAccount.GetId()
			break
		}
	}
	s.Require().NotEmpty(testPDSServiceAccountID, "PDS service account %s not found.", env.pdsServiceAccountName)
	s.controlPlane.TestPDSServiceAccountID = testPDSServiceAccountID
}

// mustHavePDStestAgentToken gets "Test PDS Service Account" and stores its Token as "Test PDS Agent Token".
func (s *PDSTestSuite) mustHavePDStestAgentToken(env environment) {
	token, resp, err := s.controlPlane.API.ServiceAccountsApi.ApiServiceAccountsIdTokenGet(s.ctx, s.controlPlane.TestPDSServiceAccountID).Execute()
	api.RequireNoError(s.T(), resp, err)
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200.")

	s.testPDSAgentToken = token.GetToken()
}

func (s *PDSTestSuite) mustHaveAuthorization(env environment) {
	var tokenSource oauth2.TokenSource
	apiToken := env.pdsToken
	if apiToken == "" {
		var err error
		tokenSource, err = auth.GetTokenSourceByPassword(
			s.ctx,
			env.controlPlane.LoginCredentials.TokenIssuerURL,
			env.controlPlane.LoginCredentials.IssuerClientID,
			env.controlPlane.LoginCredentials.IssuerClientSecret,
			env.controlPlane.LoginCredentials.Username,
			env.controlPlane.LoginCredentials.Password,
		)
		s.Require().NoError(err, "Cannot create token source")
	} else {
		tokenSource = auth.GetTokenSourceByToken(apiToken)
	}

	s.tokenSource = tokenSource
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
}

func (s *PDSTestSuite) mustHavePrometheusClient(env environment) {
	promAPI, err := prometheus.NewClient(env.controlPlane.PrometheusAPI, s.controlPlane.TestPDSTenantID, s.tokenSource)
	s.Require().NoError(err)

	s.prometheusClient = promAPI
}

func (s *PDSTestSuite) mustHaveTargetCluster(env environment) {
	tc, err := targetcluster.NewTargetCluster(s.ctx, env.targetKubeconfig)
	s.Require().NoError(err, "Cannot create target cluster.")
	s.targetCluster = tc
}

func (s *PDSTestSuite) mustInstallAgent(env environment) {
	provider, err := helminstaller.NewHelmProvider()
	s.Require().NoError(err, "Cannot create agent installer provider.")

	pdsChartConfig := helminstaller.NewPDSChartConfig(s.pdsHelmChartVersion, s.controlPlane.TestPDSTenantID, s.testPDSAgentToken, env.controlPlane.ControlPlaneAPI, env.pdsDeploymentTargetName)

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

func (s *PDSTestSuite) mustEnsureStatefulSetReady(t *testing.T, deploymentID string) {
	deployment, resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := s.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, waiterDeploymentStatusHealthyTimeout, waiterRetryInterval, func(t tests.T) {
		set, err := s.targetCluster.GetStatefulSet(s.ctx, namespace, deployment.GetClusterResourceName())
		require.NoErrorf(t, err, "Getting statefulSet for deployment %s.", deployment.GetClusterResourceName())
		require.Equalf(t, *set.Spec.Replicas, set.Status.ReadyReplicas, "Insufficient ReadyReplicas for deployment %s.", deployment.GetClusterResourceName())
	})
}

func (s *PDSTestSuite) mustEnsureLoadBalancerServicesReady(t *testing.T, deploymentID string) {
	deployment, resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := s.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, waiterLoadBalancerServicesReady, waiterRetryInterval, func(t tests.T) {
		svcs, err := s.targetCluster.ListServices(s.ctx, namespace, map[string]string{
			"name": deployment.GetClusterResourceName(),
		})
		require.NoErrorf(t, err, "Listing services for deployment %s.", deployment.GetClusterResourceName())

		for _, svc := range svcs.Items {
			if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
				ingress := svc.Status.LoadBalancer.Ingress
				require.NotEqualf(t, 0, len(ingress),
					"External ingress for service %s of deployment %s not assigned.",
					svc.GetClusterName(), deployment.GetClusterResourceName())
			}
		}
	})
}

func (s *PDSTestSuite) mustEnsureLoadBalancerHostsAccessibleIfNeeded(t *testing.T, deploymentID string) {
	deployment, resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := s.controlPlane.API.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)
	dataServiceType := dataService.GetName()

	if !s.loadBalancerAddressRequiredForTest(dataServiceType) {
		// Data service doesn't need load balancer addresses to be ready -> return.
		return
	}

	namespaceModel, resp, err := s.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)
	namespace := namespaceModel.GetName()

	// Collect all CNAME hostnames from DNSEndpoints.
	hostnames, err := s.targetCluster.GetDNSEndpoints(s.ctx, namespace, deployment.GetClusterResourceName(), "CNAME")
	require.NoError(t, err)

	// Wait until all hosts are accessible (DNS server returns an IP address for all hosts).
	if len(hostnames) > 0 {
		wait.For(t, waiterAllHostsAvailableTimeout, waiterRetryInterval, func(t tests.T) {
			dnsIPs := s.mustFlushDNSCache(t)
			jobNameSuffix := time.Now().Format("0405") // mmss
			jobName := s.mustRunHostCheckJob(t, namespace, deployment.GetClusterResourceName(), jobNameSuffix, hostnames, dnsIPs)
			s.mustWaitForJobSuccess(t, namespace, jobName)
		})
	}
}

func (s *PDSTestSuite) loadBalancerAddressRequiredForTest(dataServiceType string) bool {
	switch dataServiceType {
	case dataservices.Kafka, dataservices.RabbitMQ, dataservices.Couchbase:
		return true
	default:
		return false
	}
}

func (s *PDSTestSuite) mustRunHostCheckJob(t tests.T, namespace string, jobNamePrefix, jobNameSuffix string, hosts, dnsIPs []string) string {
	jobName := fmt.Sprintf("%s-hostcheck-%s", jobNamePrefix, jobNameSuffix)
	image := "portworx/dnsutils"
	env := []corev1.EnvVar{{
		Name:  "HOSTS",
		Value: strings.Join(hosts, " "),
	}, {
		Name:  "DNS_IPS",
		Value: strings.Join(dnsIPs, " "),
	}}
	cmd := []string{
		"/bin/bash",
		"-c",
		"for D in $DNS_IPS; do echo \"Checking on DNS $D:\"; for H in $HOSTS; do IP=$(dig +short @$D $H 2>/dev/null | head -n1); if [ -z \"$IP\" ]; then echo \"  $H - MISSING IP\";  exit 1; else echo \"  $H $IP - OK\"; fi; done; done",
	}

	job, err := s.targetCluster.CreateJob(s.ctx, namespace, jobName, image, env, cmd)
	require.NoErrorf(t, err, "Creating job %s/%s on target cluster.", namespace, jobName)
	return job.GetName()
}

func (s *PDSTestSuite) mustWaitForJobSuccess(t tests.T, namespace, jobName string) {
	// 1. Wait for the job to finish.
	s.mustWaitForJobToFinish(t, namespace, jobName, waiterHostCheckFinishedTimeout, waiterShortRetryInterval)

	// 2. Check the result.
	job, err := s.targetCluster.GetJob(s.ctx, namespace, jobName)
	require.NoErrorf(t, err, "Getting job %s/%s from target cluster.", namespace, jobName)
	require.Greaterf(t, job.Status.Succeeded, 0, "Job %s/%s did not succeed.", namespace, jobName)
}

func (s *PDSTestSuite) mustWaitForJobToFinish(t tests.T, namespace string, jobName string, timeout time.Duration, tick time.Duration) {
	wait.For(t, timeout, tick, func(t tests.T) {
		job, err := s.targetCluster.GetJob(s.ctx, namespace, jobName)
		require.NoErrorf(t, err, "Getting %s/%s job from target cluster.", namespace, jobName)
		require.Truef(t,
			job.Status.Succeeded > 0 || job.Status.Failed > 0,
			"Job did not finish (Succeeded: %d, Failed: %d)", job.Status.Succeeded, job.Status.Failed,
		)
	})
}

func (s *PDSTestSuite) mustEnsureStatefulSetReadyAndUpdatedReplicas(t *testing.T, deploymentID string) {
	deployment, resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := s.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, waiterStatefulSetReadyAndUpdatedReplicas, waiterRetryInterval, func(t tests.T) {
		set, err := s.targetCluster.GetStatefulSet(s.ctx, namespace, deployment.GetClusterResourceName())
		require.NoErrorf(t, err, "Getting statefulSet for deployment %s.", deployment.GetClusterResourceName())
		require.Equalf(t, *deployment.NodeCount, set.Status.ReadyReplicas, "ReadyReplicas don't match desired NodeCount.")
		// Also check the UpdatedReplicas count, so we are sure that all nodes were restarted after the change.
		require.Equalf(t, *deployment.NodeCount, set.Status.UpdatedReplicas, "UpdatedReplicas don't match desired NodeCount.")
	})
}

func (s *PDSTestSuite) mustEnsureStatefulSetImage(t *testing.T, deploymentID, imageTag string) {
	deployment, resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := s.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := s.controlPlane.API.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	wait.For(t, waiterDeploymentStatusHealthyTimeout, waiterRetryInterval, func(t tests.T) {
		set, err := s.targetCluster.GetStatefulSet(s.ctx, namespace, deployment.GetClusterResourceName())
		require.NoErrorf(t, err, "Getting statefulSet for deployment %s.", deployment.GetClusterResourceName())

		image, err := getDatabaseImage(dataService.GetName(), set)
		require.NoErrorf(t, err, "Getting database image of deployment %s.", deployment.GetClusterResourceName())

		require.Contains(t, image, imageTag, "StatefulSet %s does not contain image tag %q.", deployment.GetClusterResourceName(), imageTag)
	})
}

func (s *PDSTestSuite) mustEnsureDeploymentInitialized(t *testing.T, deploymentID string) {
	deployment, resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := s.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()
	clusterInitJobName := fmt.Sprintf("%s-cluster-init", deployment.GetClusterResourceName())
	nodeInitJobName := fmt.Sprintf("%s-node-init", deployment.GetClusterResourceName())

	wait.For(t, waiterDeploymentStatusHealthyTimeout, waiterRetryInterval, func(t tests.T) {
		clusterInitJob, err := s.targetCluster.GetJob(s.ctx, namespace, clusterInitJobName)
		require.NoErrorf(t, err, "Getting clusterInitJob %s/%s for deployment %s.", namespace, clusterInitJobName, deploymentID)
		require.Truef(t, isJobSucceeded(clusterInitJob), "CluterInitJob %s/%s for deployment %s not successful.", namespace, clusterInitJobName, deploymentID)

		nodeInitJob, err := s.targetCluster.GetJob(s.ctx, namespace, nodeInitJobName)
		require.NoErrorf(t, err, "Getting nodeInitJob %s/%s for deployment %s.", namespace, nodeInitJobName, deploymentID)
		require.Truef(t, isJobSucceeded(clusterInitJob), "NodeInitJob %s/%s for deployment %s not successful.", namespace, nodeInitJob, deploymentID)
	})
}

func (s *PDSTestSuite) createS3BackupTarget(backupCredentialsID, bucket, region string) (*pds.ModelsBackupTarget, *http.Response, error) {
	tenantID := s.controlPlane.TestPDSTenantID
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-s3-%s", nameSuffix)

	requestBody := pds.ControllersCreateTenantBackupTarget{
		Name:                &name,
		BackupCredentialsId: &backupCredentialsID,
		Bucket:              &bucket,
		Region:              &region,
		Type:                pointer.String("s3"),
	}
	return s.controlPlane.API.BackupTargetsApi.ApiTenantsIdBackupTargetsPost(s.ctx, tenantID).Body(requestBody).Execute()
}

func (s *PDSTestSuite) mustCreateS3BackupTarget(t *testing.T, backupCredentialsID, bucket, region string) *pds.ModelsBackupTarget {
	backupTarget, resp, err := s.createS3BackupTarget(backupCredentialsID, bucket, region)
	api.RequireNoError(t, resp, err)
	return backupTarget
}

func (s *PDSTestSuite) mustEnsureBackupTargetCreatedInTC(t *testing.T, backupTargetID, deploymentTargetID string) {
	s.mustWaitForBackupTargetState(t, backupTargetID, deploymentTargetID, "successful")
}

func (s *PDSTestSuite) mustWaitForBackupTargetState(t *testing.T, backupTargetID, deploymentTargetID, expectedFinalState string) {
	wait.For(t, waiterBackupTargetSyncedTimeout, waiterShortRetryInterval, func(t tests.T) {
		backupTargetState := s.mustGetBackupTargetState(t, backupTargetID, deploymentTargetID)
		require.Equalf(t, expectedFinalState, backupTargetState.GetState(),
			"Backup target %s failed to end up in %s state to deployment target %s.", backupTargetID, expectedFinalState, deploymentTargetID)
	})
}

func (s *PDSTestSuite) mustGetBackupTargetState(t tests.T, backupTargetID, deploymentTargetID string) pds.ModelsBackupTargetState {
	backupTargetStates, resp, err := s.controlPlane.API.BackupTargetsApi.ApiBackupTargetsIdStatesGet(s.ctx, backupTargetID).Execute()
	api.RequireNoError(t, resp, err)

	for _, backupTargetState := range backupTargetStates.GetData() {
		if backupTargetState.GetDeploymentTargetId() == deploymentTargetID {
			return backupTargetState
		}
	}
	require.Fail(t, "Backup target state for backup target %s and deployment target %s was not found.", backupTargetID, deploymentTargetID)
	return pds.ModelsBackupTargetState{}
}

func (s *PDSTestSuite) mustDeleteBackupTarget(t *testing.T, backupTargetID string) {
	// The force=true parameter ensures that all the associated backup target states are deleted even if api-workers fail
	// to delete the PX cloud credentials. This query parameter is used by default in the UI.
	resp, err := s.controlPlane.API.BackupTargetsApi.ApiBackupTargetsIdDelete(s.ctx, backupTargetID).Force("true").Execute()
	api.RequireNoError(t, resp, err)
	wait.For(t, waiterBackupStatusSucceededTimeout, waiterShortRetryInterval, func(t tests.T) {
		_, resp, err := s.controlPlane.API.BackupTargetsApi.ApiBackupTargetsIdGet(s.ctx, backupTargetID).Execute()
		assert.Error(t, err)
		assert.NotNil(t, resp)
		require.Equalf(t, http.StatusNotFound, resp.StatusCode, "Backup target %s is not deleted.", backupTargetID)
	})
}

func (s *PDSTestSuite) deleteBackupTargetIfExists(backupTargetID string) {
	// The force=true parameter ensures that all the associated backup target states are deleted even if api-workers fail
	// to delete the PX cloud credentials. This query parameter is used by default in the UI.
	resp, err := s.controlPlane.API.BackupTargetsApi.ApiBackupTargetsIdDelete(s.ctx, backupTargetID).Force("true").Execute()
	if resp.StatusCode == http.StatusNotFound {
		return
	}
	api.NoError(s.T(), resp, err)

	wait.For(s.T(), waiterBackupStatusSucceededTimeout, waiterShortRetryInterval, func(t tests.T) {
		_, resp, err := s.controlPlane.API.BackupTargetsApi.ApiBackupTargetsIdGet(s.ctx, backupTargetID).Execute()
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.Equalf(t, http.StatusNotFound, resp.StatusCode, "Backup target %s is not deleted.", backupTargetID)
	})
}

func (s *PDSTestSuite) mustEnsureBackupSuccessful(t *testing.T, deploymentID, backupName string) {
	deployment, resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	namespaceModel, resp, err := s.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	namespace := namespaceModel.GetName()

	// 1. Wait for the backup to finish.
	wait.For(t, waiterBackupStatusSucceededTimeout, waiterRetryInterval, func(t tests.T) {
		pdsBackup, err := s.targetCluster.GetPDSBackup(s.ctx, namespace, backupName)
		require.NoErrorf(t, err, "Getting backup %s/%s for deployment %s from target cluster.", namespace, backupName, deploymentID)
		require.Truef(t, isBackupFinished(pdsBackup), "Backup %s for the deployment %s did not finish.", backupName, deploymentID)
	})

	// 2. Check the result.
	pdsBackup, err := s.targetCluster.GetPDSBackup(s.ctx, namespace, backupName)
	require.NoError(t, err)

	if isBackupFailed(pdsBackup) {
		// Backup failed.
		backupJobs := pdsBackup.Status.BackupJobs
		var backupJobName string
		if len(backupJobs) > 0 {
			backupJobName = backupJobs[0].Name
		}
		logs, err := s.targetCluster.GetJobLogs(s.ctx, namespace, backupJobName, s.startTime)
		if err != nil {
			require.Fail(t, fmt.Sprintf("Backup '%s' failed.", backupName))
		} else {
			require.Fail(t, fmt.Sprintf("Backup job '%s' failed. See job logs for more details:", backupJobName), logs)
		}
	}
	require.True(t, isBackupSucceeded(pdsBackup))
}

func (s *PDSTestSuite) mustRunBasicSmokeTest(t *testing.T, deploymentID string) {
	s.mustRunLoadTestJob(t, deploymentID)
}

func (s *PDSTestSuite) mustRunLoadTestJob(t *testing.T, deploymentID string) {
	jobNamespace, jobName := s.mustCreateLoadTestJob(t, deploymentID)
	s.mustEnsureLoadTestJobSucceeded(t, jobNamespace, jobName)
	s.mustEnsureLoadTestJobLogsDoNotContain(t, jobNamespace, jobName, "ERROR|FATAL")
}

func (s *PDSTestSuite) mustCreateLoadTestJob(t *testing.T, deploymentID string) (string, string) {
	deployment, resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)
	deploymentName := deployment.GetClusterResourceName()

	namespace, resp, err := s.controlPlane.API.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := s.controlPlane.API.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)
	dataServiceType := dataService.GetName()

	dsImage, resp, err := s.controlPlane.API.ImagesApi.ApiImagesIdGet(s.ctx, deployment.GetImageId()).Execute()
	api.RequireNoError(t, resp, err)
	dsImageCreatedAt := dsImage.GetCreatedAt()

	jobName := fmt.Sprintf("%s-loadtest-%d", deployment.GetClusterResourceName(), time.Now().Unix())

	image, err := s.mustGetLoadTestJobImage(dataServiceType)
	require.NoError(t, err)

	env := s.mustGetLoadTestJobEnv(t, dataService, dsImageCreatedAt, deploymentName, namespace.GetName(), deployment.NodeCount)

	job, err := s.targetCluster.CreateJob(s.ctx, namespace.GetName(), jobName, image, env, nil)
	require.NoError(t, err)

	return namespace.GetName(), job.GetName()
}

func (s *PDSTestSuite) mustEnsureLoadTestJobSucceeded(t *testing.T, namespace, jobName string) {
	// 1. Wait for the job to finish.
	s.mustWaitForJobToFinish(t, namespace, jobName, waiterLoadTestJobFinishedTimeout, waiterShortRetryInterval)

	// 2. Check the result.
	job, err := s.targetCluster.GetJob(s.ctx, namespace, jobName)
	require.NoError(t, err)

	if job.Status.Failed > 0 {
		// Job failed.
		logs, err := s.targetCluster.GetJobLogs(s.ctx, namespace, jobName, s.startTime)
		if err != nil {
			require.Fail(t, fmt.Sprintf("Job '%s' failed.", jobName))
		} else {
			require.Fail(t, fmt.Sprintf("Job '%s' failed. See job logs for more details:", jobName), logs)
		}
	}
	require.True(t, job.Status.Succeeded > 0)
}

func (s *PDSTestSuite) mustEnsureLoadTestJobLogsDoNotContain(t *testing.T, namespace, jobName, rePattern string) {
	logs, err := s.targetCluster.GetJobLogs(s.ctx, namespace, jobName, s.startTime)
	require.NoError(t, err)
	re := regexp.MustCompile(rePattern)
	require.Nil(t, re.FindStringIndex(logs), "Job log '%s' contains pattern '%s':\n%s", jobName, rePattern, logs)
}

func (s *PDSTestSuite) mustGetLoadTestJobImage(dataServiceType string) (string, error) {
	switch dataServiceType {
	case dataservices.Cassandra:
		return "portworx/pds-loadtests:cassandra-0.0.5", nil
	case dataservices.Couchbase:
		return "portworx/pds-loadtests:couchbase-0.0.3", nil
	case dataservices.Redis:
		return "portworx/pds-loadtests:redis-0.0.3", nil
	case dataservices.ZooKeeper:
		return "portworx/pds-loadtests:zookeeper-0.0.2", nil
	case dataservices.Kafka:
		return "portworx/pds-loadtests:kafka-0.0.3", nil
	case dataservices.RabbitMQ:
		return "portworx/pds-loadtests:rabbitmq-0.0.2", nil
	case dataservices.MongoDB:
		return "portworx/pds-loadtests:mongodb-0.0.1", nil
	case dataservices.MySQL:
		return "portworx/pds-loadtests:mysql-0.0.3", nil
	case dataservices.ElasticSearch:
		return "portworx/pds-loadtests:elasticsearch-0.0.2", nil
	case dataservices.Consul:
		return "portworx/pds-loadtests:consul-0.0.1", nil
	case dataservices.Postgres:
		return "portworx/pds-loadtests:postgresql-0.0.3", nil
	default:
		return "", fmt.Errorf("loadtest job image not found for data service %s", dataServiceType)
	}
}

func (s *PDSTestSuite) mustGetLoadTestJobEnv(t *testing.T, dataService *pds.ModelsDataService, dsImageCreatedAt, deploymentName, namespace string, nodeCount *int32) []corev1.EnvVar {
	host := fmt.Sprintf("%s-%s", deploymentName, namespace)
	password := s.mustGetDBPassword(t, namespace, deploymentName)
	env := []corev1.EnvVar{
		{
			Name:  "HOST",
			Value: host,
		}, {
			Name:  "PASSWORD",
			Value: password,
		}, {
			Name:  "ITERATIONS",
			Value: "1",
		}, {
			Name:  "FAIL_ON_ERROR",
			Value: "true",
		}}

	dataServiceType := dataService.GetName()
	switch dataServiceType {
	case dataservices.Redis:
		var clusterMode string
		if nodeCount != nil && *nodeCount > 1 {
			clusterMode = "true"
		} else {
			clusterMode = "false"
		}
		var user = "pds"
		if dsImageCreatedAt != "" {
			dsCreatedAt, err := time.Parse(pdsAPITimeFormat, dsImageCreatedAt)
			if err == nil && dsCreatedAt.Before(pdsUserInRedisIntroducedAt) {
				// Older images before this change: https://github.com/portworx/pds-images-redis/pull/61 had "default" user.
				user = "default"
			}
		}
		env = append(env,
			corev1.EnvVar{
				Name:  "PDS_USER",
				Value: user,
			},
			corev1.EnvVar{
				Name:  "CLUSTER_MODE",
				Value: clusterMode,
			},
		)
	}

	return env
}

func (s *PDSTestSuite) mustRemoveDeployment(t *testing.T, deploymentID string) {
	resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdDelete(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)
}

func (s *PDSTestSuite) mustFlushDNSCache(t tests.T) []string {
	// Restarts CoreDNS pods to flush DNS cache:
	// kubectl delete pods -l k8s-app=kube-dns -n kube-system
	namespace := "kube-system"
	selector := map[string]string{"k8s-app": "kube-dns"}
	err := s.targetCluster.DeletePodsBySelector(s.ctx, namespace, selector)
	require.NoError(t, err, "Failed to delete CoreDNS pods")

	// Wait for CoreDNS pods to be fully restarted.
	wait.For(t, waiterCoreDNSRestartedTimeout, waiterShortRetryInterval, func(t tests.T) {
		set, err := s.targetCluster.ListDeployments(s.ctx, namespace, selector)
		require.NoError(t, err, "Listing CoreDNS deployments from target cluster.")
		require.Len(t, set.Items, 1, "Expected a single CoreDNS deployment.")

		deployment := set.Items[0]
		replicas := deployment.Status.Replicas
		require.Equalf(t, replicas, deployment.Status.ReadyReplicas, "Not all replicas of deployment %s are ready.", deployment.ClusterName)
		require.Equalf(t, replicas, deployment.Status.UpdatedReplicas, "Not all replicas of deployment %s are updated.", deployment.ClusterName)
	})

	// Get and return new CoreDNS pod IPs.
	pods, err := s.targetCluster.ListPods(s.ctx, namespace, selector)
	require.NoError(t, err, "Failed to get CoreDNS pods")
	var newPodIPs []string
	for _, pod := range pods.Items {
		if len(pod.Status.PodIP) > 0 && pod.Status.ContainerStatuses[0].Ready {
			newPodIPs = append(newPodIPs, pod.Status.PodIP)
		}
	}
	return newPodIPs
}

func (s *PDSTestSuite) waitForDeploymentRemoved(t *testing.T, deploymentID string) {
	wait.For(t, waiterDeploymentStatusRemovedTimeout, waiterRetryInterval, func(t tests.T) {
		_, resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
		assert.Error(t, err)
		assert.NotNil(t, resp)
		require.Equalf(t, http.StatusNotFound, resp.StatusCode, "Deployment %s is not removed.", deploymentID)
	})
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

func (s *PDSTestSuite) mustGetDBPassword(t *testing.T, namespace, deploymentName string) string {
	secretName := fmt.Sprintf("%s-creds", deploymentName)
	secret, err := s.targetCluster.GetSecret(s.ctx, namespace, secretName)
	require.NoError(t, err)

	return string(secret.Data["password"])
}

func getDatabaseImage(deploymentType string, set *appsv1.StatefulSet) (string, error) {
	var containerName string
	switch deploymentType {
	case dataservices.Postgres:
		containerName = "postgresql"
	case dataservices.Cassandra:
		containerName = "cassandra"
	case dataservices.Couchbase:
		containerName = "couchbase"
	case dataservices.Redis:
		containerName = "redis"
	case dataservices.ZooKeeper:
		containerName = "zookeeper"
	case dataservices.Kafka:
		containerName = "kafka"
	case dataservices.RabbitMQ:
		containerName = "rabbitmq"
	case dataservices.MongoDB:
		containerName = "mongos"
	case dataservices.MySQL:
		containerName = "mysql"
	case dataservices.ElasticSearch:
		containerName = "elasticsearch"
	case dataservices.Consul:
		containerName = "consul"
	default:
		return "", fmt.Errorf("unknown database type: %s", deploymentType)
	}

	for _, container := range set.Spec.Template.Spec.Containers {
		if container.Name != containerName {
			continue
		}

		return container.Image, nil
	}

	return "", fmt.Errorf("database type: %s: container %q is not found", deploymentType, containerName)
}

func (s *PDSTestSuite) deletePods(t *testing.T, deploymentID string) {
	m := map[string]string{"pds/deployment-id": deploymentID}
	err := s.targetCluster.DeletePodsBySelector(s.ctx, defaultPDSNamespaceName, m)
	require.NoError(t, err, "Cannot delete pods.")
}

func (s *PDSTestSuite) mustVerifyMetrics(t *testing.T, deploymentID string) {
	deployment, resp, err := s.controlPlane.API.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	dataService, resp, err := s.controlPlane.API.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(t, resp, err)
	dataServiceType := dataService.GetName()

	require.Contains(t, dataServiceExpectedMetrics, dataServiceType, "%s data service has no defined expected metrics")
	expectedMetrics := dataServiceExpectedMetrics[dataServiceType]

	var missingMetrics []model.LabelValue
	for _, expectedMetric := range expectedMetrics {
		// Add deployment ID to the metric label filter.
		pdsDeploymentIDMatch := parser.MustLabelMatcher(labels.MatchEqual, "pds_deployment_id", deploymentID)
		expectedMetric.LabelMatchers = append(expectedMetric.LabelMatchers, pdsDeploymentIDMatch)

		queryResult, _, err := s.prometheusClient.Query(s.ctx, expectedMetric.String(), time.Now())
		require.NoError(t, err, "prometheus: query error")

		require.IsType(t, model.Vector{}, queryResult, "prometheus: wrong result model")
		queryResultMetrics := queryResult.(model.Vector)

		if len(queryResultMetrics) == 0 {
			missingMetrics = append(missingMetrics, model.LabelValue(expectedMetric.Name))
		}
	}

	require.Equalf(t, len(missingMetrics), 0, "%s: prometheus missing %d/%d metrics: %v", dataServiceType, len(missingMetrics), len(expectedMetrics), missingMetrics)
}
