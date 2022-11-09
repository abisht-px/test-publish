package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"

	agent_installer "github.com/portworx/pds-integration-test/internal/agent-installer"
	"github.com/portworx/pds-integration-test/internal/loadgen"
	"github.com/portworx/pds-integration-test/internal/loadgen/postgresql"
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
	waiterDeploymentStatusRemovedTimeout         = time.Second * 300
	waiterLoadTestJobFinishedTimeout             = time.Second * 300
	waiterHostCheckFinishedTimeout               = time.Second * 60
	waiterAllHostsAvailableTimeout               = time.Second * 300
	waiterCoreDNSRestartedTimeout                = time.Second * 30

	pdsAPITimeFormat = "2006-01-02T15:04:05.999999Z"
)

var (
	namePrefix                 = fmt.Sprintf("integration-test-%d", time.Now().Unix())
	pdsUserInRedisIntroducedAt = time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC)
)

type applicationTemplatesInfo struct {
	AppConfigTemplateID   string
	AppConfigTemplateName string
	ResourceTemplateID    string
	ResourceTemplateName  string
}

type PDSTestSuite struct {
	suite.Suite
	ctx       context.Context
	startTime time.Time

	targetCluster              *cluster.TargetCluster
	targetClusterKubeconfig    string
	apiClient                  *pds.APIClient
	pdsAgentInstallable        agent_installer.Installable
	testPDSAccountID           string
	testPDSTenantID            string
	testPDSProjectID           string
	testPDSNamespaceID         string
	testPDSDeploymentTargetID  string
	testPDSServiceAccountID    string
	testPDSBackupCredentialsID string
	testPDSBackupTargetID      string
	testPDSAgentToken          string
	testPDSStorageTemplateID   string
	testPDSStorageTemplateName string
	testPDSTemplatesMap        map[string]applicationTemplatesInfo
	shortDeploymentSpecMap     map[PDSDeploymentSpecID]ShortDeploymentSpec
	imageVersionSpecList       []PDSImageReferenceSpec
}

func TestPDSSuite(t *testing.T) {
	suite.Run(t, new(PDSTestSuite))
}

func (s *PDSTestSuite) SetupSuite() {
	s.startTime = time.Now()
	s.ctx = context.Background()

	// Perform basic setup with sanity checks.
	env := mustHaveEnvVariables(s.T())
	s.targetClusterKubeconfig = env.targetKubeconfig
	s.shortDeploymentSpecMap = mustLoadShortDeploymentSpecMap(s.T())
	s.mustHaveTargetCluster(env)
	s.mustHaveTargetClusterNamespaces(env)
	s.mustHaveAPIClient(env)
	s.mustHavePDStestAccount(env)
	s.mustHavePDStestTenant(env)
	s.mustHavePDStestProject(env)
	s.mustHavePDStestServiceAccount(env)
	s.mustHavePDStestAgentToken(env)

	s.mustLoadImageVersions()
	s.mustInstallAgent(env)
	s.mustHavePDStestDeploymentTarget(env)
	s.mustHavePDStestNamespace(env)
	s.mustCreateApplicationTemplates()
	s.mustCreateStorageOptions()
	s.mustEnsureS3BackupCredentials(env)
	s.mustEnsureS3BackupTarget(env)
}

func (s *PDSTestSuite) TearDownSuite() {
	env := mustHaveEnvVariables(s.T())
	s.mustDeleteBackupTarget()
	s.mustDeleteBackupCredentials()
	s.mustDeleteApplicationTemplates()
	s.mustDeleteStorageOptions()
	s.mustUninstallAgent(env)
	s.mustDeletePDStestDeploymentTarget()
	if s.T().Failed() {
		s.targetCluster.LogComponents(s.T(), s.ctx, s.startTime)
	}
}

// mustHavePDStestAccount finds PDS account with name set in envrionment and stores its ID as "Test PDS Account".
func (s *PDSTestSuite) mustHavePDStestAccount(env environment) {
	// TODO: Use account name query filters
	accounts, resp, err := s.apiClient.AccountsApi.ApiAccountsGet(s.ctx).Execute()
	s.Require().NoError(err, "Error calling PDS API AccountsGet.")
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200")
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
	s.Require().NoError(err, "Error calling PDS API AccountsIdTenantsGet.")
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200")
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
	s.Require().NoError(err, "Error calling PDS API TenantsIdProjectsGet.")
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200")
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
	s.Require().Eventually(
		func() bool {
			s.testPDSDeploymentTargetID, err = getDeploymentTargetIDByName(s.ctx, s.apiClient, s.testPDSTenantID, env.pdsDeploymentTargetName)
			return err == nil
		},
		waiterDeploymentTargetNameExistsTimeout, waiterRetryInterval,
		"PDS deployment target %q does not exist: %v.", env.pdsDeploymentTargetName, err,
	)

	s.Require().Eventually(
		func() bool { return isDeploymentTargetHealthy(s.ctx, s.apiClient, s.testPDSDeploymentTargetID) },
		waiterDeploymentTargetStatusHealthyTimeout, waiterRetryInterval,
		"PDS deployment target %q is not healthy.", s.testPDSDeploymentTargetID,
	)
}

func (s *PDSTestSuite) mustDeletePDStestDeploymentTarget() {
	s.Require().Eventually(
		func() bool { return !isDeploymentTargetHealthy(s.ctx, s.apiClient, s.testPDSDeploymentTargetID) },
		waiterDeploymentTargetStatusUnhealthyTimeout, waiterRetryInterval,
		"PDS deployment target %s is still healthy.", s.testPDSDeploymentTargetID,
	)
	httpRes, err := s.apiClient.DeploymentTargetsApi.ApiDeploymentTargetsIdDelete(s.ctx, s.testPDSDeploymentTargetID).Execute()
	if err != nil {
		rawbody, parseErr := ioutil.ReadAll(httpRes.Body)
		s.Require().NoError(parseErr, "Parse api error")
		s.Require().NoError(err, "Error calling PDS API DeploymentTargetsIdDelete: %s", rawbody)
	}
	s.Require().Equal(204, httpRes.StatusCode, "PDS API must return HTTP 204")
}

func (s *PDSTestSuite) mustHavePDStestNamespace(env environment) {
	s.Require().Eventually(
		func() bool {
			var err error
			s.testPDSNamespaceID, err = getNamespaceIDByName(s.ctx, s.apiClient, s.testPDSDeploymentTargetID, env.pdsNamespaceName)
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
	s.Require().NoError(err, "Error calling PDS API TenantsIdServiceAccountsGet.")
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200")
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
	s.Require().NoError(err, "Error calling PDS API ServiceAccountsIdTokenGet.")
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200.")

	s.testPDSAgentToken = token.GetToken()
}

func (s *PDSTestSuite) mustHaveAPIClient(env environment) {
	endpointUrl, err := url.Parse(env.controlPlaneAPI)
	s.Require().NoError(err, "Cannot parse control plane URL.")

	apiConf := pds.NewConfiguration()
	apiConf.Host = endpointUrl.Host
	apiConf.Scheme = endpointUrl.Scheme

	bearerToken, err := auth.GetBearerToken(s.ctx,
		env.secrets.tokenIssuerURL,
		env.secrets.issuerClientID,
		env.secrets.issuerClientSecret,
		env.secrets.pdsUsername,
		env.secrets.pdsPassword,
	)
	s.Require().NoError(err, "Cannot get bearer token.")

	s.ctx = context.WithValue(s.ctx,
		pds.ContextAPIKeys,
		map[string]pds.APIKey{
			"ApiKeyAuth": {Key: bearerToken, Prefix: "Bearer"},
		})

	s.apiClient = pds.NewAPIClient(apiConf)
}

func (s *PDSTestSuite) mustHaveTargetCluster(env environment) {
	tc, err := cluster.NewTargetCluster(env.targetKubeconfig)
	s.Require().NoError(err, "Cannot create target cluster.")
	s.targetCluster = tc
}

func (s *PDSTestSuite) mustInstallAgent(env environment) {
	provider, err := agent_installer.NewHelmProvider()
	s.Require().NoError(err, "Cannot create agent installer provider.")

	helmSelectorAgent14, err := agent_installer.NewSelectorHelmPDS18WithName(env.targetKubeconfig, s.testPDSTenantID, s.testPDSAgentToken, env.controlPlaneAPI, env.pdsDeploymentTargetName)
	s.Require().NoError(err, "Cannot create agent installer selector.")

	installer, err := provider.Installer(helmSelectorAgent14)
	s.Require().NoError(err, "Cannot get agent installer for version selector %s.", helmSelectorAgent14.ConstraintsString())

	err = installer.Install(s.ctx)
	s.Require().NoError(err, "Cannot install agent for version %s selector.", helmSelectorAgent14.ConstraintsString())
	s.pdsAgentInstallable = installer
}

func (s *PDSTestSuite) mustUninstallAgent(env environment) {
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
	err = s.targetCluster.DeleteDetachedPXVolumes(s.ctx, env.pxNamespaceName)
	s.NoError(err, "Cannot delete detached PX volumes.")
	err = s.targetCluster.DeletePXCredentials(s.ctx, env.pxNamespaceName)
	s.NoError(err, "Cannot delete detached PX volumes.")

}

func (s *PDSTestSuite) mustLoadImageVersions() {
	imageVersions, err := getAllImageVersions(s.ctx, s.apiClient)
	s.Require().NoError(err, "Error while reading image versions.")
	s.Require().NotEmpty(imageVersions, "No image versions found.")
	s.imageVersionSpecList = imageVersions
}

func (s *PDSTestSuite) mustDeployDeploymentSpec(deployment ShortDeploymentSpec) string {
	image := findImageVersionForRecord(&deployment, s.imageVersionSpecList)
	s.Require().NotNil(image, "No image found for deployment %s.", deployment.ServiceName)

	s.setDeploymentDefaults(&deployment)

	deploymentID, err := createPDSDeployment(s.ctx, s.apiClient, &deployment, image, s.testPDSTenantID, s.testPDSDeploymentTargetID, s.testPDSProjectID, s.testPDSNamespaceID)
	s.Require().NoError(err, "Error while creating deployment %s.", deployment.ServiceName)
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
	dsTemplate, found := s.testPDSTemplatesMap[deployment.ServiceName]
	if found {
		if deployment.ResourceSettingsTemplateName == "" {
			deployment.ResourceSettingsTemplateName = dsTemplate.ResourceTemplateName
		}
		if deployment.AppConfigTemplateName == "" {
			deployment.AppConfigTemplateName = dsTemplate.AppConfigTemplateName
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

	_, httpRes, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdPut(s.ctx, deploymentID).Body(req).Execute()
	if err != nil {
		rawbody, parseErr := ioutil.ReadAll(httpRes.Body)
		s.Require().NoError(parseErr, "Parse api error")
		s.Require().NoError(err, "Error calling PDS API DeploymentTargetsIdPut: %s", rawbody)
	}
	s.Require().NoError(err, "Update %s deployment.", deploymentID)
}

func (s *PDSTestSuite) mustEnsureDeploymentHealthy(deploymentID string) {
	s.Require().Eventually(
		func() bool {
			return isDeploymentHealthy(s.ctx, s.apiClient, deploymentID)
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Deployment %s is not healthy.", deploymentID,
	)
}

func (s *PDSTestSuite) mustEnsureStatefulSetReady(deploymentID string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

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
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

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
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)
	dataService, _, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	s.Require().NoError(err)
	dataServiceType := dataService.GetName()

	if !s.loadBalancerAddressRequiredForTest(dataServiceType) {
		// Data service doesn't need load balancer addresses to be ready -> return.
		return
	}

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)
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
	case dbKafka, dbRabbitMQ:
		return true
	default:
		return false
	}
}

func (s *PDSTestSuite) mustRunHostCheckJob(namespace string, jobNamePrefix, jobNameSuffix string, hosts, dnsIPs []string) string {

	jobName := fmt.Sprintf("%s-hostcheck-%s", jobNamePrefix, jobNameSuffix)
	image := "tutum/dnsutils" // TODO DS-3289: Copy this image into portworx repo.
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

func (s *PDSTestSuite) mustEnsureStatefulSetReadyAndUpdatedReplicas(deploymentID string, replicas int) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	namespace := namespaceModel.GetName()
	s.Require().Eventually(
		func() bool {
			set, err := s.targetCluster.GetStatefulSet(s.ctx, namespace, deployment.GetClusterResourceName())
			if err != nil {
				return false
			}

			// Also check the UpdatedReplicas count, so we are sure that all nodes were restarted after the change.
			return int(set.Status.ReadyReplicas) == replicas && int(set.Status.UpdatedReplicas) == replicas
		},
		waiterStatefulSetReadyAndUpdatedReplicas, waiterRetryInterval,
		"Deployment %s is expected to have %d ready and updated replicas.", deploymentID, replicas,
	)
}

func (s *PDSTestSuite) mustEnsureStatefulSetImage(deploymentID, imageTag string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	dataService, _, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	s.Require().NoError(err)

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
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

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

func (s *PDSTestSuite) mustCreateBackup(deploymentID string) *pds.ModelsBackup {
	backupReq := pds.ControllersCreateDeploymentBackup{
		BackupLevel:    pointer.String("snapshot"),
		BackupTargetId: pointer.String(s.testPDSBackupTargetID),
		BackupType:     pointer.String("adhoc"),
	}
	backup, _, err := s.apiClient.BackupsApi.ApiDeploymentsIdBackupsPost(s.ctx, deploymentID).Body(backupReq).Execute()
	s.Require().NoError(err)

	return backup
}

func (s *PDSTestSuite) mustDeleteBackup(backupID string) {
	_, err := s.apiClient.BackupsApi.ApiBackupsIdDelete(s.ctx, backupID).Execute()
	s.Require().NoError(err)
}

func (s *PDSTestSuite) mustEnsureS3BackupCredentials(env environment) {
	tenantID := s.testPDSTenantID
	backupCredentialsName := fmt.Sprintf("%s-backupcredentials-s3", namePrefix)

	backupCredentialsResp, _, err := s.apiClient.BackupCredentialsApi.ApiTenantsIdBackupCredentialsGet(s.ctx, tenantID).Execute()
	s.Require().NoError(err)

	for _, backupCredentials := range backupCredentialsResp.Data {
		if backupCredentials.GetName() == backupCredentialsName {
			s.testPDSBackupCredentialsID = backupCredentials.GetId()
			break
		}
	}

	if s.testPDSBackupCredentialsID == "" {
		backupCredentials := pds.ControllersCreateBackupCredentialsRequest{
			Name: &backupCredentialsName,
			Credentials: &pds.ControllersCredentials{
				S3: &pds.ModelsS3Credentials{
					Endpoint:  pointer.String(env.backupTarget.credentials.s3.endpoint),
					AccessKey: pointer.String(env.backupTarget.credentials.s3.accessKey),
					SecretKey: pointer.String(env.backupTarget.credentials.s3.secretKey),
				},
			},
		}
		backupCreds, _, err := s.apiClient.BackupCredentialsApi.ApiTenantsIdBackupCredentialsPost(s.ctx, tenantID).Body(backupCredentials).Execute()
		s.Require().NoError(err)

		s.testPDSBackupCredentialsID = backupCreds.GetId()
	}
}

func (s *PDSTestSuite) mustEnsureS3BackupTarget(env environment) {
	tenantID := s.testPDSTenantID
	backupTargetName := fmt.Sprintf("%s-backuptarget-s3", namePrefix)

	backupTargets, _, err := s.apiClient.BackupTargetsApi.ApiTenantsIdBackupTargetsGet(s.ctx, tenantID).Execute()
	s.Require().NoError(err)

	for _, backupTarget := range backupTargets.Data {
		if backupTarget.GetName() == backupTargetName {
			s.testPDSBackupTargetID = backupTarget.GetId()
			break
		}
	}

	if s.testPDSBackupTargetID == "" {
		backupTarget := pds.ControllersCreateTenantBackupTarget{
			Name:                pointer.StringPtr(backupTargetName),
			BackupCredentialsId: pointer.StringPtr(s.testPDSBackupCredentialsID),
			Bucket:              pointer.String(env.backupTarget.bucket),
			Region:              pointer.String(env.backupTarget.region),
			Type:                pointer.String("s3"),
		}
		backupTargetModel, _, err := s.apiClient.BackupTargetsApi.ApiTenantsIdBackupTargetsPost(s.ctx, tenantID).Body(backupTarget).Execute()
		s.Require().NoError(err)

		s.testPDSBackupTargetID = backupTargetModel.GetId()
	}
}

func (s *PDSTestSuite) mustDeleteBackupCredentials() {
	if s.testPDSBackupCredentialsID != "" {
		_, err := s.apiClient.BackupCredentialsApi.ApiBackupCredentialsIdDelete(s.ctx, s.testPDSBackupCredentialsID).Execute()
		s.NoError(err)
	}
}

func (s *PDSTestSuite) mustDeleteBackupTarget() {
	if s.testPDSBackupTargetID != "" {
		_, err := s.apiClient.BackupTargetsApi.ApiBackupTargetsIdDelete(s.ctx, s.testPDSBackupTargetID).Execute()
		s.NoError(err)

		s.Require().Eventually(
			func() bool {
				_, httpResp, err := s.apiClient.BackupTargetsApi.ApiBackupTargetsIdGet(s.ctx, s.testPDSBackupTargetID).Execute()
				return err != nil && httpResp != nil && httpResp.StatusCode == http.StatusNotFound
			},
			waiterBackupStatusSucceededTimeout, waiterRetryInterval,
			"Backup target %s is not deleted.", s.testPDSBackupTargetID,
		)
	}
}

func (s *PDSTestSuite) mustCreateStorageOptions() {
	storageTemplate := pds.ControllersCreateStorageOptionsTemplatesRequest{
		Name:   pointer.StringPtr(namePrefix),
		Repl:   pointer.Int32Ptr(1),
		Secure: pointer.BoolPtr(false),
		Fs:     pointer.StringPtr("xfs"),
		Fg:     pointer.BoolPtr(false),
	}
	storageTemplateResp, _, err := s.apiClient.StorageOptionsTemplatesApi.
		ApiTenantsIdStorageOptionsTemplatesPost(s.ctx, s.testPDSTenantID).
		Body(storageTemplate).Execute()
	s.Require().NoError(err)

	s.testPDSStorageTemplateID = storageTemplateResp.GetId()
	s.testPDSStorageTemplateName = storageTemplateResp.GetName()
}

func (s *PDSTestSuite) mustCreateApplicationTemplates() {
	dataServicesTemplates := make(map[string]applicationTemplatesInfo)
	for _, imageVersion := range s.imageVersionSpecList {
		dsTemplate, found := dataServiceTemplatesSpec[imageVersion.ServiceName]
		if !found {
			continue
		}
		_, found = dataServicesTemplates[imageVersion.ServiceName]
		if found {
			continue
		}

		configTemplateBody := dsTemplate.configurationTemplate
		configTemplateBody.Name = pointer.StringPtr(namePrefix)
		configTemplateBody.DataServiceId = pds.PtrString(imageVersion.DataServiceID)

		configTemplate, _, err := s.apiClient.ApplicationConfigurationTemplatesApi.
			ApiTenantsIdApplicationConfigurationTemplatesPost(s.ctx, s.testPDSTenantID).
			Body(configTemplateBody).Execute()
		s.Require().NoError(err, "data service: %s", imageVersion.ServiceName)

		resourceTemplateBody := dsTemplate.resourceTemplate
		resourceTemplateBody.Name = pointer.StringPtr(namePrefix)
		resourceTemplateBody.DataServiceId = pds.PtrString(imageVersion.DataServiceID)

		resourceTemplate, _, err := s.apiClient.ResourceSettingsTemplatesApi.
			ApiTenantsIdResourceSettingsTemplatesPost(s.ctx, s.testPDSTenantID).
			Body(resourceTemplateBody).Execute()
		s.Require().NoError(err, "data service: %s", imageVersion.ServiceName)

		dataServicesTemplates[imageVersion.ServiceName] = applicationTemplatesInfo{
			AppConfigTemplateID:   configTemplate.GetId(),
			AppConfigTemplateName: configTemplate.GetName(),
			ResourceTemplateID:    resourceTemplate.GetId(),
			ResourceTemplateName:  resourceTemplate.GetName(),
		}
	}
	s.testPDSTemplatesMap = dataServicesTemplates
}

func (s *PDSTestSuite) mustDeleteStorageOptions() {
	_, err := s.apiClient.StorageOptionsTemplatesApi.ApiStorageOptionsTemplatesIdDelete(s.ctx, s.testPDSStorageTemplateID).Execute()
	s.Require().NoError(err)
}

func (s *PDSTestSuite) mustDeleteApplicationTemplates() {
	for _, dsTemplate := range s.testPDSTemplatesMap {
		appConfigTemplateID := dsTemplate.AppConfigTemplateID
		_, err := s.apiClient.ApplicationConfigurationTemplatesApi.ApiApplicationConfigurationTemplatesIdDelete(s.ctx, appConfigTemplateID).Execute()
		s.Require().NoError(err)

		resourceTemplateID := dsTemplate.ResourceTemplateID
		_, err = s.apiClient.ResourceSettingsTemplatesApi.ApiResourceSettingsTemplatesIdDelete(s.ctx, resourceTemplateID).Execute()
		s.Require().NoError(err)
	}
}

func (s *PDSTestSuite) mustEnsureBackupSuccessful(deploymentID, backupName string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

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
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	dataService, _, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	s.Require().NoError(err)
	dataServiceType := dataService.GetName()

	switch dataServiceType {
	case dbPostgres:
		s.mustReadWriteData(deploymentID)
	default:
		s.mustRunLoadTestJob(deploymentID)
	}
}

func (s *PDSTestSuite) mustReadWriteData(deploymentID string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespace, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	dataService, _, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	s.Require().NoError(err)

	port, err := getDefaultDBPort(dataService.GetName())
	s.Require().NoError(err)

	podName := s.mustGetDeploymentPodName(deploymentID, deployment.GetClusterResourceName(), dataService.GetName(), namespace.GetName())
	pf, err := s.targetCluster.PortforwardPod(namespace.GetName(), podName, port)
	s.Require().NoError(err)

	dbPassword := s.mustGetDBPassword(namespace.GetName(), deployment.GetClusterResourceName())
	loadtest, err := s.mustGetLoadgenFor(dataService.GetName(), "localhost", strconv.Itoa(pf.Local), dbPassword)
	s.Require().NoError(err)

	err = loadtest.Run(s.ctx)
	s.Require().NoError(err)
}

func (s *PDSTestSuite) mustRunLoadTestJob(deploymentID string) {
	jobNamespace, jobName := s.mustCreateLoadTestJob(deploymentID)
	s.mustEnsureLoadTestJobSucceeded(jobNamespace, jobName)
	s.mustEnsureLoadTestJobLogsDoNotContain(jobNamespace, jobName, "ERROR|FATAL")
}

func (s *PDSTestSuite) mustCreateLoadTestJob(deploymentID string) (string, string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)
	deploymentName := deployment.GetClusterResourceName()

	namespace, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	dataService, _, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	s.Require().NoError(err)
	dataServiceType := dataService.GetName()

	jobName := fmt.Sprintf("%s-loadtest", deployment.GetClusterResourceName())

	image, err := s.mustGetLoadTestJobImage(dataServiceType)
	s.Require().NoError(err)

	env := s.mustGetLoadTestJobEnv(dataService, deploymentName, namespace.GetName(), deployment.NodeCount)

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
	case dbRedis:
		return "portworx/pds-loadtests:redis-0.0.3", nil
	case dbZooKeeper:
		return "portworx/pds-loadtests:zookeeper-0.0.2", nil
	case dbKafka:
		return "portworx/pds-loadtests:kafka-0.0.3", nil
	case dbRabbitMQ:
		return "portworx/pds-loadtests:rabbitmq-0.0.2", nil
	case dbMySQL:
		return "portworx/pds-loadtests:mysql-0.0.3", nil
	default:
		return "", fmt.Errorf("loadtest job image not found for data service %s", dataServiceType)
	}
}

func (s *PDSTestSuite) mustGetLoadTestJobEnv(dataService *pds.ModelsDataService, deploymentName, namespace string, nodeCount *int32) []corev1.EnvVar {
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
		if dataService.CreatedAt != nil {
			dsCreatedAt, err := time.Parse(pdsAPITimeFormat, *dataService.CreatedAt)
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
	_, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdDelete(s.ctx, deploymentID).Execute()
	s.Require().NoError(err, "Error while removing deployment %s.", deploymentID)
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
			_, httpResp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
			return httpResp != nil && httpResp.StatusCode == 404 && err != nil
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

func (s *PDSTestSuite) mustGetLoadgenFor(dataServiceType, host, port, dbPassword string) (loadgen.Interface, error) {
	switch dataServiceType {
	case dbPostgres:
		return s.newPostgresLoadgen(host, port, "pds", dbPassword), nil
	}

	return nil, fmt.Errorf("loadgen for the %s database is not found", dataServiceType)
}

func (s *PDSTestSuite) newPostgresLoadgen(host, port, dbUser, dbPassword string) loadgen.Interface {
	postgresLoader, err := postgresql.New(postgresql.Config{
		User:     dbUser,
		Password: dbPassword,
		Host:     host,
		Port:     port,
		Count:    5,
		Logger:   newTestLogger(s.T()),
	})
	s.Require().NoError(err)

	return postgresLoader
}

func (s *PDSTestSuite) mustGetDeploymentPodName(deploymentID, deploymentName, deploymentType, namespace string) string {
	switch deploymentType {
	case dbPostgres:
		// For the port-forward tunnel needs the 'master' pod otherwise we'll get the 'read-only' error.
		selector := map[string]string{
			"pds/deployment-id": deploymentID,
			"role":              "master",
		}
		podList, err := s.targetCluster.ListPods(s.ctx, namespace, selector)
		s.Require().NoError(err)
		s.Require().NotEmpty(podList.Items, "get pods for the label selector: %s", labels.FormatLabels(selector))
		return podList.Items[0].Name
	default:
		return getDeploymentPodName(deploymentName)
	}
}

func getDefaultDBPort(deploymentType string) (int, error) {
	switch deploymentType {
	case dbPostgres:
		return 5432, nil
	}
	return -1, fmt.Errorf("has no default port for the %s database", deploymentType)
}

func getDatabaseImage(deploymentType string, set *appsv1.StatefulSet) (string, error) {
	var containerName string
	switch deploymentType {
	case dbPostgres:
		containerName = "postgresql"
	case dbCassandra:
		containerName = "cassandra"
	case dbRedis:
		containerName = "redis"
	case dbZooKeeper:
		containerName = "zookeeper"
	case dbKafka:
		containerName = "kafka"
	case dbRabbitMQ:
		containerName = "rabbitmq"
	case dbMySQL:
		containerName = "mysql"
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
