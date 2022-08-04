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

type PDSTestSuite struct {
	suite.Suite
	ctx       context.Context
	startTime time.Time

	targetCluster           *cluster.TargetCluster
	apiClient               *pds.APIClient
	pdsAgentInstallable     agent_installer.Installable
	testPDSAccountID        string
	testPDSTenantID         string
	testPDSServiceAccountID string
	testPDSAgentToken       string
}

func TestPDSSuite(t *testing.T) {
	suite.Run(t, new(PDSTestSuite))
}

func (s *PDSTestSuite) SetupSuite() {
	s.startTime = time.Now()
	s.ctx = context.Background()

	// Perform basic setup with sanity checks.
	env := mustHaveEnvVariables(s.T())
	s.mustHaveTargetCluster(env)
	s.mustHaveAPIClient(env)
	s.mustHavePDStestAccount(env)
	s.mustHavePDStestTenant(env)
	s.mustHavePDStestServiceAccount(env)
	s.mustHavePDStestAgentToken(env)

	//TODO: Pass agent install values from external configuration source.
	s.mustInstallAgent(env)
}

func (s *PDSTestSuite) TearDownSuite() {
	s.mustUninstallAgent()
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
