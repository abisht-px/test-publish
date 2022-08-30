package test

import (
	"context"
	"net/url"
	"testing"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/suite"

	agent_installer "github.com/portworx/pds-integration-test/internal/agent-installer"
	"github.com/portworx/pds-integration-test/test/auth"
	cluster "github.com/portworx/pds-integration-test/test/cluster"
)

const (
	waiterRetryInterval                          = time.Second * 10
	waiterDeploymentTargetNameExistsTimeout      = time.Second * 30
	waiterNamespaceExistsTimeout                 = time.Second * 30
	waiterDeploymentTargetStatusHealthyTimeout   = time.Second * 120
	waiterDeploymentTargetStatusUnhealthyTimeout = time.Second * 300
	waiterDeploymentStatusHealthyTimeout         = time.Second * 300
	waiterDeploymentStatusRemovedTimeout         = time.Second * 300
)

type PDSTestSuite struct {
	suite.Suite
	ctx       context.Context
	startTime time.Time

	targetCluster             *cluster.TargetCluster
	apiClient                 *pds.APIClient
	pdsAgentInstallable       agent_installer.Installable
	testPDSAccountID          string
	testPDSTenantID           string
	testPDSProjectID          string
	testPDSNamespaceID        string
	testPDSDeploymentTargetID string
	testPDSServiceAccountID   string
	testPDSAgentToken         string
	shortDeploymentSpecMap    map[PDSDeploymentSpecID]ShortDeploymentSpec
	imageVersionSpecList      []PDSImageReferenceSpec
}

func TestPDSSuite(t *testing.T) {
	suite.Run(t, new(PDSTestSuite))
}

func (s *PDSTestSuite) SetupSuite() {
	s.startTime = time.Now()
	s.ctx = context.Background()

	// Perform basic setup with sanity checks.
	env := mustHaveEnvVariables(s.T())
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
}

func (s *PDSTestSuite) TearDownSuite() {
	s.mustUninstallAgent()
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
	s.Eventually(
		func() bool {
			var err error
			s.testPDSDeploymentTargetID, err = getDeploymentTargetIDByName(s.ctx, s.apiClient, s.testPDSTenantID, env.pdsDeploymentTargetName)
			return err == nil
		},
		waiterDeploymentTargetNameExistsTimeout, waiterRetryInterval,
		"PDS deployment target %s does not exist.", env.pdsDeploymentTargetName,
	)

	s.Eventually(
		func() bool { return isDeploymentTargetHealthy(s.ctx, s.apiClient, s.testPDSDeploymentTargetID) },
		waiterDeploymentTargetStatusHealthyTimeout, waiterRetryInterval,
		"PDS deployment target %s is not healthy.", s.testPDSDeploymentTargetID,
	)
}

func (s *PDSTestSuite) mustDeletePDStestDeploymentTarget() {
	s.Eventually(
		func() bool { return !isDeploymentTargetHealthy(s.ctx, s.apiClient, s.testPDSDeploymentTargetID) },
		waiterDeploymentTargetStatusUnhealthyTimeout, waiterRetryInterval,
		"PDS deployment target %s is still healthy.", s.testPDSDeploymentTargetID,
	)
	httRes, err := s.apiClient.DeploymentTargetsApi.ApiDeploymentTargetsIdDelete(s.ctx, s.testPDSDeploymentTargetID).Execute()
	s.Require().NoError(err, "Error calling PDS API DeploymentTargetsIdDelete.")
	s.Require().Equal(204, httRes.StatusCode, "PDS API must return HTTP 204")
}

func (s *PDSTestSuite) mustHavePDStestNamespace(env environment) {
	s.Eventually(
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

	helmSelectorAgent14, err := agent_installer.NewSelectorHelmPDS14WithName(env.targetKubeconfig, s.testPDSTenantID, s.testPDSAgentToken, env.controlPlaneAPI, env.pdsDeploymentTargetName)
	s.Require().NoError(err, "Cannot create agent installer selector.")

	installer, err := provider.Installer(helmSelectorAgent14)
	s.Require().NoError(err, "Cannot get agent installer for version selector %s.", helmSelectorAgent14.ConstraintsString())

	err = installer.Install(s.ctx)
	s.Require().NoError(err, "Cannot install agent for version %s selector.", helmSelectorAgent14.ConstraintsString())
	s.pdsAgentInstallable = installer
}

func (s *PDSTestSuite) mustUninstallAgent() {
	err := s.pdsAgentInstallable.Uninstall(s.ctx)
	s.Require().NoError(err)
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

	deploymentID, err := createPDSDeployment(s.ctx, s.apiClient, &deployment, image, s.testPDSTenantID, s.testPDSDeploymentTargetID, s.testPDSProjectID, s.testPDSNamespaceID)
	s.Require().NoError(err, "Error while creating deployment %s.", deployment.ServiceName)
	s.Require().NotEmpty(deploymentID, "Deployment ID is empty.")

	return deploymentID
}

func (s *PDSTestSuite) mustEnsureDeploymentHealty(deploymentID string) {
	s.Eventually(
		func() bool {
			return isDeploymentHealthy(s.ctx, s.apiClient, deploymentID)
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Deployment %s is not healthy.", deploymentID,
	)
}

func (s *PDSTestSuite) mustRemoveDeployment(deploymentID string) {
	_, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdDelete(s.ctx, deploymentID).Execute()
	s.Require().NoError(err, "Error while removing deployment %s.", deploymentID)
}

func (s *PDSTestSuite) mustEnsureDeploymentRemoved(deploymentID string) {
	s.Eventually(
		func() bool {
			_, httpResp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
			return httpResp.StatusCode == 404 && err != nil
		},
		waiterDeploymentStatusRemovedTimeout, waiterRetryInterval,
		"Deployment %s is not removed.", deploymentID,
	)
}

func (s *PDSTestSuite) mustHaveTargetClusterNamespaces(env environment) {
	nss := []string{env.pdsNamespaceName}
	s.targetCluster.EnsureNamespaces(s.T(), s.ctx, nss)
}
