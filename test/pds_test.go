package test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/helminstaller"
	"github.com/portworx/pds-integration-test/internal/random"
	"github.com/portworx/pds-integration-test/test/api"
	"github.com/portworx/pds-integration-test/test/auth"
	"github.com/portworx/pds-integration-test/test/cluster"
)

const (
	waiterShortRetryInterval                     = time.Second * 5
	waiterRetryInterval                          = time.Second * 10
	waiterDeploymentTargetNameExistsTimeout      = time.Second * 30
	waiterNamespaceExistsTimeout                 = time.Second * 30
	waiterDeploymentTargetStatusHealthyTimeout   = time.Second * 120
	waiterDeploymentTargetStatusUnhealthyTimeout = time.Second * 300
	waiterDeploymentStatusHealthyTimeout         = time.Second * 300
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
)

// Info for a single template.
type templateInfo struct {
	ID   string
	Name string
}

// Info for all app config and resource templates which belong to a data service.
type dataServiceTemplateInfo struct {
	AppConfigTemplates []templateInfo
	ResourceTemplates  []templateInfo
}

type PDSTestSuite struct {
	suite.Suite
	ctx       context.Context
	startTime time.Time

	targetCluster              *cluster.TargetCluster
	targetClusterKubeconfig    string
	apiClient                  *pds.APIClient
	pdsAgentInstallable        *helminstaller.InstallableHelmPDS
	pdsHelmChartVersion        string
	testPDSAccountID           string
	testPDSTenantID            string
	testPDSProjectID           string
	testPDSNamespaceID         string
	testPDSDeploymentTargetID  string
	testPDSServiceAccountID    string
	testPDSAgentToken          string
	testPDSStorageTemplateID   string
	testPDSStorageTemplateName string
	testPDSTemplatesMap        map[string]dataServiceTemplateInfo
	config                     environment
	imageVersionSpecList       []PDSImageReferenceSpec
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
	s.targetClusterKubeconfig = env.targetKubeconfig
	s.mustHaveTargetCluster(env)
	s.mustHaveTargetClusterNamespaces(env)
	s.mustHaveAPIClient(env)
	s.mustHavePDSMetadata(env)
	s.mustHavePDStestAccount(env)
	s.mustHavePDStestTenant(env)
	s.mustHavePDStestProject(env)
	s.mustLoadImageVersions()
	if shouldInstallPDSHelmChart(s.pdsHelmChartVersion) {
		s.mustHavePDStestServiceAccount(env)
		s.mustHavePDStestAgentToken(env)
		s.mustInstallAgent(env)
	}
	s.mustHavePDStestDeploymentTarget(env)
	s.mustHavePDStestNamespace(env)
	s.mustCreateApplicationTemplates()
	s.mustCreateStorageOptions()
}

func (s *PDSTestSuite) TearDownSuite() {
	env := mustHaveEnvVariables(s.T())
	// Do not fail fast on cleanups - we want to clean up as much as possible even on failures.
	s.deleteApplicationTemplates()
	s.deleteStorageOptions()
	if shouldInstallPDSHelmChart(env.pdsHelmChartVersion) {
		s.uninstallAgent(env)
		s.deletePDStestDeploymentTarget()
	}
	if s.T().Failed() {
		s.targetCluster.LogComponents(s.T(), s.ctx, s.startTime)
	}
}

func (s *PDSTestSuite) mustCreateUserAPIKey(
	ctx context.Context, apiClient *pds.APIClient, expiresAt time.Time, name string) *pds.ModelsUserAPIKey {
	expirationDate := expiresAt.Format(time.RFC3339)
	requestBody := pds.RequestsCreateUserAPIKeyRequest{
		ExpiresAt: &expirationDate,
		Name:      &name,
	}
	userApiKey, response, err := apiClient.UserAPIKeyApi.ApiUserApiKeyPost(ctx).Body(requestBody).Execute()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(response.StatusCode, http.StatusCreated, "user api key was not created as expected")

	return userApiKey
}

// mustHavePDSMetadata gets PDS API metadata and stores the PDS helm chart version in the test suite.
func (s *PDSTestSuite) mustHavePDSMetadata(env environment) {
	metadata, resp, err := s.apiClient.MetadataApi.ApiMetadataGet(s.ctx).Execute()
	api.RequireNoError(s.T(), resp, err)

	// If user didn't specify the helm chart version, let's use the one configured in PDS API.
	if env.pdsHelmChartVersion == "" {
		s.pdsHelmChartVersion = metadata.GetHelmChartVersion()
	} else {
		s.pdsHelmChartVersion = env.pdsHelmChartVersion
	}
}

// mustHavePDStestAccount finds PDS account with name set in envrionment and stores its ID as "Test PDS Account".
func (s *PDSTestSuite) mustHavePDStestAccount(env environment) {
	// TODO: Use account name query filters
	accounts, resp, err := s.apiClient.AccountsApi.ApiAccountsGet(s.ctx).Execute()
	api.RequireNoError(s.T(), resp, err)
	s.Require().NotEmpty(accounts, "PDS API must return at least one account.")

	var testPDSAccountID string
	for _, account := range accounts.GetData() {
		if account.GetName() == env.pdsAccountName {
			testPDSAccountID = account.GetId()
			break
		}
	}
	s.Require().NotEmpty(testPDSAccountID, "PDS account %s not found.", env.pdsAccountName)
	s.testPDSAccountID = testPDSAccountID
}

// mustHavePDStestTenant finds PDS tenant in Test PDS Account with name set in environment and stores its ID as "Test PDS Tenant".
func (s *PDSTestSuite) mustHavePDStestTenant(env environment) {
	// TODO: Use tenant name query filters
	tenants, resp, err := s.apiClient.TenantsApi.ApiAccountsIdTenantsGet(s.ctx, s.testPDSAccountID).Execute()
	api.RequireNoError(s.T(), resp, err)
	s.Require().NotEmpty(tenants, "PDS API must return at least one tenant.")

	var testPDSTenantID string
	for _, tenant := range tenants.GetData() {
		if tenant.GetName() == env.pdsTenantName {
			testPDSTenantID = tenant.GetId()
			break
		}
	}
	s.Require().NotEmpty(testPDSTenantID, "PDS tenant %s not found.", env.pdsTenantName)
	s.testPDSTenantID = testPDSTenantID
}

// mustHavePDStestProject finds PDS project in Test PDS Tenant with name set in environment and stores its ID as "Test PDS Project".
func (s *PDSTestSuite) mustHavePDStestProject(env environment) {
	// TODO: Use project name query filters
	projects, resp, err := s.apiClient.ProjectsApi.ApiTenantsIdProjectsGet(s.ctx, s.testPDSTenantID).Execute()
	api.RequireNoError(s.T(), resp, err)
	s.Require().NotEmpty(projects, "PDS API must return at least one project.")

	var testPDSProjectID string
	for _, project := range projects.GetData() {
		if project.GetName() == env.pdsProjectName {
			testPDSProjectID = project.GetId()
			break
		}
	}
	s.Require().NotEmpty(testPDSProjectID, "PDS project %s not found.", env.pdsProjectName)
	s.testPDSProjectID = testPDSProjectID
}

func (s *PDSTestSuite) mustHavePDStestDeploymentTarget(env environment) {
	var err error
	s.requireNowOrEventually(
		func() bool {
			s.testPDSDeploymentTargetID, err = getDeploymentTargetIDByName(s.T(), s.ctx, s.apiClient, s.testPDSTenantID, env.pdsDeploymentTargetName)
			return err == nil
		},
		waiterDeploymentTargetNameExistsTimeout, waiterRetryInterval,
		"PDS deployment target %q does not exist: %v.", env.pdsDeploymentTargetName, err,
	)

	s.requireNowOrEventually(
		func() bool { return isDeploymentTargetHealthy(s.T(), s.ctx, s.apiClient, s.testPDSDeploymentTargetID) },
		waiterDeploymentTargetStatusHealthyTimeout, waiterRetryInterval,
		"PDS deployment target %q is not healthy.", s.testPDSDeploymentTargetID,
	)
}

func (s *PDSTestSuite) deletePDStestDeploymentTarget() {
	s.nowOrEventually(
		func() bool { return !isDeploymentTargetHealthy(s.T(), s.ctx, s.apiClient, s.testPDSDeploymentTargetID) },
		waiterDeploymentTargetStatusUnhealthyTimeout, waiterRetryInterval,
		"PDS deployment target %s is still healthy.", s.testPDSDeploymentTargetID,
	)
	resp, err := s.apiClient.DeploymentTargetsApi.ApiDeploymentTargetsIdDelete(s.ctx, s.testPDSDeploymentTargetID).Execute()
	api.NoErrorf(s.T(), resp, err, "Deleting deployment target %s.", s.testPDSDeploymentTargetID)
	s.Equal(http.StatusNoContent, resp.StatusCode, "Unexpected response code from deleting deployment target.")
}

func (s *PDSTestSuite) mustHavePDStestNamespace(env environment) {
	s.requireNowOrEventually(
		func() bool {
			var err error
			s.testPDSNamespaceID, err = getNamespaceIDByName(s.T(), s.ctx, s.apiClient, s.testPDSDeploymentTargetID, env.pdsNamespaceName)
			return err == nil
		},
		waiterNamespaceExistsTimeout, waiterRetryInterval,
		"PDS Namespace %s does not exist.", env.pdsNamespaceName,
	)
}

// mustHavePDStestServiceAccount finds PDS Service account in Test PDS tenant with name set in environment and stores its ID as "Test PDS Service Account".
func (s *PDSTestSuite) mustHavePDStestServiceAccount(env environment) {
	// TODO: Use service account name query filters
	serviceAccounts, resp, err := s.apiClient.ServiceAccountsApi.ApiTenantsIdServiceAccountsGet(s.ctx, s.testPDSTenantID).Execute()
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
	s.testPDSServiceAccountID = testPDSServiceAccountID
}

// mustHavePDStestAgentToken gets "Test PDS Service Account" and stores its Token as "Test PDS Agent Token".
func (s *PDSTestSuite) mustHavePDStestAgentToken(env environment) {
	token, resp, err := s.apiClient.ServiceAccountsApi.ApiServiceAccountsIdTokenGet(s.ctx, s.testPDSServiceAccountID).Execute()
	api.RequireNoError(s.T(), resp, err)
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200.")

	s.testPDSAgentToken = token.GetToken()
}

func (s *PDSTestSuite) mustHaveAPIClient(env environment) {
	endpointUrl, err := url.Parse(env.controlPlaneAPI)
	s.Require().NoError(err, "Cannot parse control plane URL.")

	apiConf := pds.NewConfiguration()
	apiConf.Host = endpointUrl.Host
	apiConf.Scheme = endpointUrl.Scheme

	bearerToken := env.pdsToken
	if bearerToken == "" {
		bearerToken, err = auth.GetBearerToken(s.ctx,
			env.secrets.tokenIssuerURL,
			env.secrets.issuerClientID,
			env.secrets.issuerClientSecret,
			env.secrets.pdsUsername,
			env.secrets.pdsPassword,
		)
		s.Require().NoError(err, "Cannot get bearer token.")
	}

	s.ctx = context.WithValue(s.ctx,
		pds.ContextAPIKeys,
		map[string]pds.APIKey{
			"ApiKeyAuth": {Key: bearerToken, Prefix: "Bearer"},
		})

	s.apiClient = pds.NewAPIClient(apiConf)
}

func (s *PDSTestSuite) mustHaveTargetCluster(env environment) {
	tc, err := cluster.NewTargetCluster(s.ctx, env.targetKubeconfig)
	s.Require().NoError(err, "Cannot create target cluster.")
	s.targetCluster = tc
}

func (s *PDSTestSuite) mustInstallAgent(env environment) {
	provider, err := helminstaller.NewHelmProvider()
	s.Require().NoError(err, "Cannot create agent installer provider.")

	pdsChartConfig := helminstaller.NewPDSChartConfig(s.pdsHelmChartVersion, s.testPDSTenantID, s.testPDSAgentToken, env.controlPlaneAPI, env.pdsDeploymentTargetName)

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
	s.nowOrEventually(func() bool {
		err := s.targetCluster.DeleteDetachedPXVolumes(s.ctx)
		return err != nil
	}, 5*time.Minute, 10*time.Second, "Cannot delete detached PX volumes.")
	err = s.targetCluster.DeletePXCloudCredentials(s.ctx)
	s.NoError(err, "Cannot delete PX cloud credentials.")
}

func (s *PDSTestSuite) mustLoadImageVersions() {
	imageVersions, err := getAllImageVersions(s.T(), s.ctx, s.apiClient)
	s.Require().NoError(err, "Error while reading image versions.")
	s.Require().NotEmpty(imageVersions, "No image versions found.")
	s.imageVersionSpecList = imageVersions
}

func (s *PDSTestSuite) mustDeployDeploymentSpec(deployment ShortDeploymentSpec) string {
	image := findImageVersionForRecord(&deployment, s.imageVersionSpecList)
	s.Require().NotNil(image, "No image found for deployment %s.", deployment.DataServiceName)

	s.setDeploymentDefaults(&deployment)

	deploymentID, err := createPDSDeployment(s.T(), s.ctx, s.apiClient, &deployment, image, s.testPDSTenantID, s.testPDSDeploymentTargetID, s.testPDSProjectID, s.testPDSNamespaceID)
	s.Require().NoError(err, "Error while creating deployment %s.", deployment.DataServiceName)
	s.Require().NotEmpty(deploymentID, "Deployment ID is empty.")

	return deploymentID
}

func (s *PDSTestSuite) setDeploymentDefaults(deployment *ShortDeploymentSpec) {
	if deployment.ServiceType == "" {
		deployment.ServiceType = "LoadBalancer"
	}
	if deployment.StorageOptionName == "" {
		deployment.StorageOptionName = s.testPDSStorageTemplateName
	}
	dsTemplates, found := s.testPDSTemplatesMap[deployment.DataServiceName]
	if found {
		if deployment.ResourceSettingsTemplateName == "" {
			deployment.ResourceSettingsTemplateName = dsTemplates.ResourceTemplates[0].Name
		}
		if deployment.AppConfigTemplateName == "" {
			deployment.AppConfigTemplateName = dsTemplates.AppConfigTemplates[0].Name
		}
	}
}

func (s *PDSTestSuite) mustUpdateDeployment(deploymentID string, spec *ShortDeploymentSpec) {
	req := pds.ControllersUpdateDeploymentRequest{}
	if spec.ImageVersionTag != "" || spec.ImageVersionBuild != "" {
		image := findImageVersionForRecord(spec, s.imageVersionSpecList)
		s.Require().NotNil(image, "Update deployment: no image found for %s version.", spec.ImageVersionTag)

		req.ImageId = &image.ImageID
	}
	if spec.NodeCount != 0 {
		nodeCount := int32(spec.NodeCount)
		req.NodeCount = &nodeCount
	}

	deployment, resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)

	if spec.ResourceSettingsTemplateName != "" {
		resourceTemplate, err := getResourceSettingsTemplateByName(s.T(), s.ctx, s.apiClient, s.testPDSTenantID, spec.ResourceSettingsTemplateName, *deployment.DataServiceId)
		s.Require().NoError(err)
		req.ResourceSettingsTemplateId = resourceTemplate.Id
	}

	if spec.AppConfigTemplateName != "" {
		appConfigTemplate, err := getAppConfigTemplateByName(s.T(), s.ctx, s.apiClient, s.testPDSTenantID, spec.AppConfigTemplateName, *deployment.DataServiceId)
		s.Require().NoError(err)
		req.ApplicationConfigurationTemplateId = appConfigTemplate.Id
	}

	_, resp, err = s.apiClient.DeploymentsApi.ApiDeploymentsIdPut(s.ctx, deploymentID).Body(req).Execute()
	api.RequireNoErrorf(s.T(), resp, err, "update %s deployment", deploymentID)
}

func (s *PDSTestSuite) mustEnsureDeploymentHealthy(deploymentID string) {
	s.Require().Eventually(
		func() bool {
			return isDeploymentHealthy(s.T(), s.ctx, s.apiClient, deploymentID)
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Deployment %s is not healthy.", deploymentID,
	)
}

func (s *PDSTestSuite) mustEnsureStatefulSetReady(deploymentID string) {
	deployment, resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespaceModel, resp, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespace := namespaceModel.GetName()
	s.Require().Eventually(
		func() bool {
			set, err := s.targetCluster.GetStatefulSet(s.ctx, namespace, deployment.GetClusterResourceName())
			if err != nil {
				return false
			}

			return *set.Spec.Replicas == set.Status.ReadyReplicas
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Deployment %s is not ready.", deploymentID,
	)
}

func (s *PDSTestSuite) mustEnsureLoadBalancerServicesReady(deploymentID string) {
	deployment, resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespaceModel, resp, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespace := namespaceModel.GetName()
	s.Require().Eventually(
		func() bool {
			svcs, err := s.targetCluster.ListServices(s.ctx, namespace, map[string]string{
				"name": deployment.GetClusterResourceName(),
			})
			if err != nil {
				return false
			}

			for _, svc := range svcs.Items {
				if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
					ingress := svc.Status.LoadBalancer.Ingress
					if len(ingress) == 0 {
						// Load balancer is not initialized yet, external address was not assigned yet.
						return false
					}
				}
			}

			return true
		},
		waiterLoadBalancerServicesReady, waiterRetryInterval,
		"Load balancers of %s are not ready.", deploymentID,
	)
}

func (s *PDSTestSuite) mustEnsureLoadBalancerHostsAccessibleIfNeeded(deploymentID string) {
	deployment, resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)

	dataService, resp, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(s.T(), resp, err)
	dataServiceType := dataService.GetName()

	if !s.loadBalancerAddressRequiredForTest(dataServiceType) {
		// Data service doesn't need load balancer addresses to be ready -> return.
		return
	}

	namespaceModel, resp, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)
	namespace := namespaceModel.GetName()

	// Collect all CNAME hostnames from DNSEndpoints.
	hostnames, err := s.targetCluster.GetDNSEndpoints(s.ctx, namespace, deployment.GetClusterResourceName(), "CNAME")
	s.Require().NoError(err)

	// Wait until all hosts are accessible (DNS server returns an IP address for all hosts).
	if len(hostnames) > 0 {
		s.Require().Eventually(
			func() bool {
				dnsIPs := s.mustFlushDNSCache()
				jobNameSuffix := time.Now().Format("0405") // mmss
				jobName := s.mustRunHostCheckJob(namespace, deployment.GetClusterResourceName(), jobNameSuffix, hostnames, dnsIPs)
				hostsAccessible := s.mustWaitForHostCheckJobResult(namespace, jobName)
				return hostsAccessible
			},
			waiterAllHostsAvailableTimeout, waiterRetryInterval,
			"Failed to wait for all hosts to be available:\n%s", strings.Join(hostnames, "\n"),
		)
	}
}

func (s *PDSTestSuite) loadBalancerAddressRequiredForTest(dataServiceType string) bool {
	switch dataServiceType {
	case dbKafka, dbRabbitMQ, dbCouchbase:
		return true
	default:
		return false
	}
}

func (s *PDSTestSuite) mustRunHostCheckJob(namespace string, jobNamePrefix, jobNameSuffix string, hosts, dnsIPs []string) string {
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
	s.Require().NoError(err)
	return job.GetName()
}

func (s *PDSTestSuite) mustWaitForHostCheckJobResult(namespace, jobName string) bool {
	// 1. Wait for the job to finish.
	s.waitForJobToFinish(namespace, jobName, waiterHostCheckFinishedTimeout, waiterShortRetryInterval)

	// 2. Check the result.
	job, err := s.targetCluster.GetJob(s.ctx, namespace, jobName)
	s.Require().NoError(err)

	return job.Status.Succeeded > 0
}

func (s *PDSTestSuite) waitForJobToFinish(namespace string, jobName string, waitFor time.Duration, tick time.Duration) {
	s.Require().Eventually(
		func() bool {
			job, err := s.targetCluster.GetJob(s.ctx, namespace, jobName)
			return err == nil && (job.Status.Succeeded > 0 || job.Status.Failed > 0)
		},
		waitFor, tick,
		"Failed to wait for job %s to finish.", jobName,
	)
}

func (s *PDSTestSuite) mustEnsureStatefulSetReadyAndUpdatedReplicas(deploymentID string) {
	deployment, resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespaceModel, resp, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespace := namespaceModel.GetName()
	s.Require().Eventually(
		func() bool {
			set, err := s.targetCluster.GetStatefulSet(s.ctx, namespace, deployment.GetClusterResourceName())
			if err != nil {
				return false
			}

			// Also check the UpdatedReplicas count, so we are sure that all nodes were restarted after the change.
			return set.Status.ReadyReplicas == *deployment.NodeCount && set.Status.UpdatedReplicas == *deployment.NodeCount
		},
		waiterStatefulSetReadyAndUpdatedReplicas, waiterRetryInterval,
		"Deployment %s is expected to have %d ready and updated replicas.", deploymentID, *deployment.NodeCount,
	)
}

func (s *PDSTestSuite) mustEnsureStatefulSetImage(deploymentID, imageTag string) {
	deployment, resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespaceModel, resp, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)

	dataService, resp, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespace := namespaceModel.GetName()
	s.Require().Eventually(
		func() bool {
			set, err := s.targetCluster.GetStatefulSet(s.ctx, namespace, deployment.GetClusterResourceName())
			if err != nil {
				return false
			}
			image, err := getDatabaseImage(dataService.GetName(), set)
			if err != nil {
				return false
			}
			return strings.Contains(image, imageTag)
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Statefulset %s is expected to have %s image tag.", deployment.GetClusterResourceName(), imageTag,
	)
}

func (s *PDSTestSuite) mustEnsureDeploymentInitialized(deploymentID string) {
	deployment, resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespaceModel, resp, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespace := namespaceModel.GetName()
	clusterInitJobName := fmt.Sprintf("%s-cluster-init", deployment.GetClusterResourceName())
	nodeInitJobName := fmt.Sprintf("%s-node-init", deployment.GetClusterResourceName())

	s.Require().Eventually(
		func() bool {
			clusterInitJob, err := s.targetCluster.GetJob(s.ctx, namespace, clusterInitJobName)
			if err != nil {
				return false
			}
			if !isJobSucceeded(clusterInitJob) {
				return false
			}

			nodeInitJob, err := s.targetCluster.GetJob(s.ctx, namespace, nodeInitJobName)
			if err != nil {
				return false
			}
			return isJobSucceeded(nodeInitJob)
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Deployment %s is not ready.", deploymentID,
	)
}

func (s *PDSTestSuite) mustCreateBackup(deploymentID, backupTargetID string) *pds.ModelsBackup {
	requestBody := pds.ControllersCreateDeploymentBackup{
		BackupLevel:    pointer.String("snapshot"),
		BackupTargetId: pointer.String(backupTargetID),
		BackupType:     pointer.String("adhoc"),
	}
	backup, resp, err := s.apiClient.BackupsApi.ApiDeploymentsIdBackupsPost(s.ctx, deploymentID).Body(requestBody).Execute()
	api.RequireNoError(s.T(), resp, err)

	return backup
}

func (s *PDSTestSuite) mustDeleteBackup(backupID string) {
	resp, err := s.apiClient.BackupsApi.ApiBackupsIdDelete(s.ctx, backupID).Execute()
	api.RequireNoError(s.T(), resp, err)
}

func (s *PDSTestSuite) mustCreateS3BackupCredentials(endpoint, accessKey, secretKey string) *pds.ModelsBackupCredentials {
	tenantID := s.testPDSTenantID
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-s3-%s", nameSuffix)

	requestBody := pds.ControllersCreateBackupCredentialsRequest{
		Name: &name,
		Credentials: &pds.ControllersCredentials{
			S3: &pds.ModelsS3Credentials{
				Endpoint:  &endpoint,
				AccessKey: &accessKey,
				SecretKey: &secretKey,
			},
		},
	}
	backupCredentials, resp, err := s.apiClient.BackupCredentialsApi.ApiTenantsIdBackupCredentialsPost(s.ctx, tenantID).Body(requestBody).Execute()
	api.RequireNoError(s.T(), resp, err)
	return backupCredentials
}

func (s *PDSTestSuite) mustCreateS3BackupTarget(backupCredentialsID, bucket, region string) *pds.ModelsBackupTarget {
	tenantID := s.testPDSTenantID
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-s3-%s", nameSuffix)

	requestBody := pds.ControllersCreateTenantBackupTarget{
		Name:                &name,
		BackupCredentialsId: &backupCredentialsID,
		Bucket:              &bucket,
		Region:              &region,
		Type:                pointer.String("s3"),
	}
	backupTarget, resp, err := s.apiClient.BackupTargetsApi.ApiTenantsIdBackupTargetsPost(s.ctx, tenantID).Body(requestBody).Execute()
	api.RequireNoError(s.T(), resp, err)

	return backupTarget
}

func (s *PDSTestSuite) mustEnsureBackupTargetSynced(backupTargetID, deploymentTargetID string) {
	s.Eventually(func() bool {
		backupTargetState := s.mustGetBackupTargetState(backupTargetID, deploymentTargetID)
		return backupTargetState.GetState() == "successful"
	}, waiterBackupTargetSyncedTimeout, waiterRetryInterval,
		"Backup target %s failed to get synced to deployment target %s", backupTargetID, deploymentTargetID,
	)
}

func (s *PDSTestSuite) mustGetBackupTargetState(backupTargetID, deploymentTargetID string) pds.ModelsBackupTargetState {
	backupTargetStates, resp, err := s.apiClient.BackupTargetsApi.ApiBackupTargetsIdStatesGet(s.ctx, backupTargetID).Execute()
	api.RequireNoError(s.T(), resp, err)

	for _, backupTargetState := range backupTargetStates.GetData() {
		if backupTargetState.GetDeploymentTargetId() == deploymentTargetID {
			return backupTargetState
		}
	}
	s.Require().Fail("Backup target state for backup target %s and deployment target %s was not found.", backupTargetID, deploymentTargetID)
	return pds.ModelsBackupTargetState{}
}

func (s *PDSTestSuite) mustDeleteBackupCredentials(backupCredentialsID string) {
	resp, err := s.apiClient.BackupCredentialsApi.ApiBackupCredentialsIdDelete(s.ctx, backupCredentialsID).Execute()
	api.RequireNoError(s.T(), resp, err)
}

func (s *PDSTestSuite) mustDeleteBackupTarget(backupTargetID string) {
	resp, err := s.apiClient.BackupTargetsApi.ApiBackupTargetsIdDelete(s.ctx, backupTargetID).Execute()
	api.RequireNoError(s.T(), resp, err)

	s.Require().Eventually(
		func() bool {
			_, resp, err := s.apiClient.BackupTargetsApi.ApiBackupTargetsIdGet(s.ctx, backupTargetID).Execute()
			return err != nil && resp != nil && resp.StatusCode == http.StatusNotFound
		},
		waiterBackupStatusSucceededTimeout, waiterRetryInterval,
		"Backup target %s is not deleted.", backupTargetID,
	)
}

func (s *PDSTestSuite) mustCreateStorageOptions() {
	storageTemplate := pds.ControllersCreateStorageOptionsTemplateRequest{
		Name:   pointer.StringPtr(namePrefix),
		Repl:   pointer.Int32Ptr(1),
		Secure: pointer.BoolPtr(false),
		Fs:     pointer.StringPtr("xfs"),
		Fg:     pointer.BoolPtr(false),
	}
	storageTemplateResp, resp, err := s.apiClient.StorageOptionsTemplatesApi.
		ApiTenantsIdStorageOptionsTemplatesPost(s.ctx, s.testPDSTenantID).
		Body(storageTemplate).Execute()
	api.RequireNoError(s.T(), resp, err)
	s.Require().NoError(err)

	s.testPDSStorageTemplateID = storageTemplateResp.GetId()
	s.testPDSStorageTemplateName = storageTemplateResp.GetName()
}

func (s *PDSTestSuite) mustCreateApplicationTemplates() {
	dataServicesTemplates := make(map[string]dataServiceTemplateInfo)
	for _, imageVersion := range s.imageVersionSpecList {
		templatesSpec, found := dataServiceTemplatesSpec[imageVersion.DataServiceName]
		if !found {
			continue
		}
		_, found = dataServicesTemplates[imageVersion.DataServiceName]
		if found {
			continue
		}

		var resultTemplateInfo dataServiceTemplateInfo
		for _, configTemplateSpec := range templatesSpec.configurationTemplates {
			configTemplateBody := configTemplateSpec
			if configTemplateBody.Name == nil {
				configTemplateBody.Name = pointer.StringPtr(namePrefix)
			}
			configTemplateBody.DataServiceId = pds.PtrString(imageVersion.DataServiceID)

			configTemplate, resp, err := s.apiClient.ApplicationConfigurationTemplatesApi.
				ApiTenantsIdApplicationConfigurationTemplatesPost(s.ctx, s.testPDSTenantID).
				Body(configTemplateBody).Execute()
			api.RequireNoError(s.T(), resp, err)

			configTemplateInfo := templateInfo{
				ID:   configTemplate.GetId(),
				Name: configTemplate.GetName(),
			}

			resultTemplateInfo.AppConfigTemplates = append(resultTemplateInfo.AppConfigTemplates, configTemplateInfo)
		}

		for _, resourceTemplateSpec := range templatesSpec.resourceTemplates {
			resourceTemplateBody := resourceTemplateSpec
			if resourceTemplateBody.Name == nil {
				resourceTemplateBody.Name = pointer.StringPtr(namePrefix)
			}
			resourceTemplateBody.DataServiceId = pds.PtrString(imageVersion.DataServiceID)

			resourceTemplate, resp, err := s.apiClient.ResourceSettingsTemplatesApi.
				ApiTenantsIdResourceSettingsTemplatesPost(s.ctx, s.testPDSTenantID).
				Body(resourceTemplateBody).Execute()
			api.RequireNoError(s.T(), resp, err)

			resourceTemplateInfo := templateInfo{
				ID:   resourceTemplate.GetId(),
				Name: resourceTemplate.GetName(),
			}

			resultTemplateInfo.ResourceTemplates = append(resultTemplateInfo.ResourceTemplates, resourceTemplateInfo)
		}

		dataServicesTemplates[imageVersion.DataServiceName] = resultTemplateInfo
	}
	s.testPDSTemplatesMap = dataServicesTemplates
}

func (s *PDSTestSuite) deleteStorageOptions() {
	resp, err := s.apiClient.StorageOptionsTemplatesApi.ApiStorageOptionsTemplatesIdDelete(s.ctx, s.testPDSStorageTemplateID).Execute()
	api.NoErrorf(s.T(), resp, err, "Deleting test storage options template (%s)", s.testPDSStorageTemplateID)
}

func (s *PDSTestSuite) deleteApplicationTemplates() {
	for _, dsTemplate := range s.testPDSTemplatesMap {
		for _, configTemplateInfo := range dsTemplate.AppConfigTemplates {
			resp, err := s.apiClient.ApplicationConfigurationTemplatesApi.ApiApplicationConfigurationTemplatesIdDelete(s.ctx, configTemplateInfo.ID).Execute()
			api.NoErrorf(s.T(), resp, err, "Deleting configuration template (ID=%s, name=%s).", configTemplateInfo.ID, configTemplateInfo.Name)
		}

		for _, resourceTemplateInfo := range dsTemplate.ResourceTemplates {
			resp, err := s.apiClient.ResourceSettingsTemplatesApi.ApiResourceSettingsTemplatesIdDelete(s.ctx, resourceTemplateInfo.ID).Execute()
			api.NoErrorf(s.T(), resp, err, "Deleting resource settings template (ID=%s, name=%s)", resourceTemplateInfo.ID, resourceTemplateInfo.Name)
		}
	}
}

func (s *PDSTestSuite) mustEnsureBackupSuccessful(deploymentID, backupName string) {
	deployment, resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespaceModel, resp, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)

	namespace := namespaceModel.GetName()

	// 1. Wait for the backup to finish.
	s.Require().Eventually(
		func() bool {
			pdsBackup, err := s.targetCluster.GetPDSBackup(s.ctx, namespace, backupName)
			return err == nil && isBackupFinished(pdsBackup)
		},
		waiterBackupStatusSucceededTimeout, waiterRetryInterval,
		"Backup %s for the %s deployment is not finished.", backupName, deploymentID,
	)

	// 2. Check the result.
	pdsBackup, err := s.targetCluster.GetPDSBackup(s.ctx, namespace, backupName)
	s.Require().NoError(err)

	if isBackupFailed(pdsBackup) {
		// Backup failed.
		backupJobs := pdsBackup.Status.BackupJobs
		var backupJobName string
		if len(backupJobs) > 0 {
			backupJobName = backupJobs[0].Name
		}
		logs, err := s.targetCluster.GetJobLogs(s.T(), s.ctx, namespace, backupJobName, s.startTime)
		if err != nil {
			s.Require().Fail(fmt.Sprintf("Backup '%s' failed.", backupName))
		} else {
			s.Require().Fail(fmt.Sprintf("Backup job '%s' failed. See job logs for more details:", backupJobName), logs)
		}
	}
	s.Require().True(isBackupSucceeded(pdsBackup))
}

func (s *PDSTestSuite) mustRunBasicSmokeTest(deploymentID string) {
	s.mustRunLoadTestJob(deploymentID)
}

func (s *PDSTestSuite) mustRunLoadTestJob(deploymentID string) {
	jobNamespace, jobName := s.mustCreateLoadTestJob(deploymentID)
	s.mustEnsureLoadTestJobSucceeded(jobNamespace, jobName)
	s.mustEnsureLoadTestJobLogsDoNotContain(jobNamespace, jobName, "ERROR|FATAL")
}

func (s *PDSTestSuite) mustCreateLoadTestJob(deploymentID string) (string, string) {
	deployment, resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
	deploymentName := deployment.GetClusterResourceName()

	namespace, resp, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	api.RequireNoError(s.T(), resp, err)

	dataService, resp, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	api.RequireNoError(s.T(), resp, err)
	dataServiceType := dataService.GetName()

	dsImage, resp, err := s.apiClient.ImagesApi.ApiImagesIdGet(s.ctx, deployment.GetImageId()).Execute()
	api.RequireNoError(s.T(), resp, err)
	dsImageCreatedAt := dsImage.GetCreatedAt()

	jobName := fmt.Sprintf("%s-loadtest-%d", deployment.GetClusterResourceName(), time.Now().Unix())

	image, err := s.mustGetLoadTestJobImage(dataServiceType)
	s.Require().NoError(err)

	env := s.mustGetLoadTestJobEnv(dataService, dsImageCreatedAt, deploymentName, namespace.GetName(), deployment.NodeCount)

	job, err := s.targetCluster.CreateJob(s.ctx, namespace.GetName(), jobName, image, env, nil)
	s.Require().NoError(err)

	return namespace.GetName(), job.GetName()
}

func (s *PDSTestSuite) mustEnsureLoadTestJobSucceeded(namespace, jobName string) {
	// 1. Wait for the job to finish.
	s.waitForJobToFinish(namespace, jobName, waiterLoadTestJobFinishedTimeout, waiterShortRetryInterval)

	// 2. Check the result.
	job, err := s.targetCluster.GetJob(s.ctx, namespace, jobName)
	s.Require().NoError(err)

	if job.Status.Failed > 0 {
		// Job failed.
		logs, err := s.targetCluster.GetJobLogs(s.T(), s.ctx, namespace, jobName, s.startTime)
		if err != nil {
			s.Require().Fail(fmt.Sprintf("Job '%s' failed.", jobName))
		} else {
			s.Require().Fail(fmt.Sprintf("Job '%s' failed. See job logs for more details:", jobName), logs)
		}
	}
	s.Require().True(job.Status.Succeeded > 0)
}

func (s *PDSTestSuite) mustEnsureLoadTestJobLogsDoNotContain(namespace, jobName, rePattern string) {
	logs, err := s.targetCluster.GetJobLogs(s.T(), s.ctx, namespace, jobName, s.startTime)
	s.Require().NoError(err)
	re := regexp.MustCompile(rePattern)
	s.Require().Nil(re.FindStringIndex(logs), "Job log '%s' contains pattern '%s':\n%s", jobName, rePattern, logs)
}

func (s *PDSTestSuite) mustGetLoadTestJobImage(dataServiceType string) (string, error) {
	switch dataServiceType {
	case dbCassandra:
		return "portworx/pds-loadtests:cassandra-0.0.3", nil
	case dbCouchbase:
		return "portworx/pds-loadtests:couchbase-0.0.3", nil
	case dbRedis:
		return "portworx/pds-loadtests:redis-0.0.3", nil
	case dbZooKeeper:
		return "portworx/pds-loadtests:zookeeper-0.0.2", nil
	case dbKafka:
		return "portworx/pds-loadtests:kafka-0.0.3", nil
	case dbRabbitMQ:
		return "portworx/pds-loadtests:rabbitmq-0.0.2", nil
	case dbMongoDB:
		return "portworx/pds-loadtests:mongodb-0.0.1", nil
	case dbMySQL:
		return "portworx/pds-loadtests:mysql-0.0.3", nil
	case dbElasticSearch:
		return "portworx/pds-loadtests:elasticsearch-0.0.2", nil
	case dbConsul:
		return "portworx/pds-loadtests:consul-0.0.1", nil
	case dbPostgres:
		return "portworx/pds-loadtests:postgresql-0.0.3", nil
	default:
		return "", fmt.Errorf("loadtest job image not found for data service %s", dataServiceType)
	}
}

func (s *PDSTestSuite) mustGetLoadTestJobEnv(dataService *pds.ModelsDataService, dsImageCreatedAt, deploymentName, namespace string, nodeCount *int32) []corev1.EnvVar {
	host := fmt.Sprintf("%s-%s", deploymentName, namespace)
	password := s.mustGetDBPassword(namespace, deploymentName)
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
	case dbRedis:
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

func (s *PDSTestSuite) mustRemoveDeployment(deploymentID string) {
	resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdDelete(s.ctx, deploymentID).Execute()
	api.RequireNoError(s.T(), resp, err)
}

func (s *PDSTestSuite) mustFlushDNSCache() []string {
	// Restarts CoreDNS pods to flush DNS cache:
	// kubectl delete pods -l k8s-app=kube-dns -n kube-system
	namespace := "kube-system"
	selector := map[string]string{"k8s-app": "kube-dns"}
	err := s.targetCluster.DeletePodsBySelector(s.ctx, namespace, selector)
	s.Require().NoError(err, "Failed to delete coredns pods")

	// Wait for CoreDNS pods to be fully restarted.
	s.Require().Eventually(
		func() bool {
			set, err := s.targetCluster.ListDeployments(s.ctx, namespace, selector)
			if err != nil || len(set.Items) != 1 {
				return false
			}

			d := set.Items[0]
			replicas := d.Status.Replicas
			return d.Status.ReadyReplicas == replicas && d.Status.UpdatedReplicas == replicas
		},
		waiterCoreDNSRestartedTimeout, waiterShortRetryInterval,
		"Failed to wait for CoreDNS pods to be restarted.",
	)

	// Get and return new CoreDNS pod IPs.
	pods, err := s.targetCluster.ListPods(s.ctx, namespace, selector)
	s.Require().NoError(err, "Failed to get CoreDNS pods")
	var newPodIPs []string
	for _, pod := range pods.Items {
		if len(pod.Status.PodIP) > 0 && pod.Status.ContainerStatuses[0].Ready {
			newPodIPs = append(newPodIPs, pod.Status.PodIP)
		}
	}
	return newPodIPs
}

func (s *PDSTestSuite) mustEnsureDeploymentRemoved(deploymentID string) {
	s.Require().Eventually(
		func() bool {
			_, resp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
			return resp != nil && resp.StatusCode == 404 && err != nil
		},
		waiterDeploymentStatusRemovedTimeout, waiterRetryInterval,
		"Deployment %s is not removed.", deploymentID,
	)
}

func (s *PDSTestSuite) mustHaveTargetClusterNamespaces(env environment) {
	nss := []string{env.pdsNamespaceName}
	s.targetCluster.EnsureNamespaces(s.T(), s.ctx, nss)
}

func (s *PDSTestSuite) mustGetDBPassword(namespace, deploymentName string) string {
	secretName := fmt.Sprintf("%s-creds", deploymentName)
	secret, err := s.targetCluster.GetSecret(s.ctx, namespace, secretName)
	s.Require().NoError(err)

	return string(secret.Data["password"])
}

func getDatabaseImage(deploymentType string, set *appsv1.StatefulSet) (string, error) {
	var containerName string
	switch deploymentType {
	case dbPostgres:
		containerName = "postgresql"
	case dbCassandra:
		containerName = "cassandra"
	case dbCouchbase:
		containerName = "couchbase"
	case dbRedis:
		containerName = "redis"
	case dbZooKeeper:
		containerName = "zookeeper"
	case dbKafka:
		containerName = "kafka"
	case dbRabbitMQ:
		containerName = "rabbitmq"
	case dbMongoDB:
		containerName = "mongos"
	case dbMySQL:
		containerName = "mysql"
	case dbElasticSearch:
		containerName = "elasticsearch"
	case dbConsul:
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

// requireNowOrEventually tries to evaluate the condition immediately, or waits for specified number of time to become truthful.
// This is useful in cases when the target cluster is already registered to a control plane -> there's no need to wait.
func (s *PDSTestSuite) requireNowOrEventually(condition func() bool, waitFor time.Duration, tick time.Duration, msgAndArgs ...interface{}) {
	if s.nowOrEventually(condition, waitFor, tick, msgAndArgs...) {
		return
	}

	s.T().FailNow()
}

// nowOrEventually tries to evaluate the condition immediately, or waits for specified number of time to become truthful.
// This is useful in cases when the target cluster is already registered to a control plane -> there's no need to wait.
func (s *PDSTestSuite) nowOrEventually(condition func() bool, waitFor time.Duration, tick time.Duration, msgAndArgs ...interface{}) bool {
	if condition() {
		return true
	}
	return s.Eventually(condition, waitFor, tick, msgAndArgs...)
}

func (s *PDSTestSuite) deletePods(deploymentID string) {
	m := map[string]string{"pds/deployment-id": deploymentID}
	err := s.targetCluster.DeletePodsBySelector(s.ctx, defaultPDSNamespaceName, m)
	s.NoError(err, "Cannot delete pods.")
}
