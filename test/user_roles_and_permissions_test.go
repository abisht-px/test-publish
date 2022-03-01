package test

import (
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (s *PDSTestSuite) TestAccountsTenantsProjectsList() {
	// Creating Accounts API client
	accounts := ListAccounts(s)
	s.Require().Equal(1, len(accounts))
	accountUUID := accounts[0].GetId()
	accountName := accounts[0].GetName()
	s.T().Logf("Account Detail- Name: %s, UUID: %s ", accountName, accountUUID)
	s.Require().Equal("Portworx", accountName, "Portworx account doesn't exist.")

	tenants := ListTenants(s, accountUUID)
	s.Require().Equal(1, len(tenants))
	tenantName := tenants[0].GetName()
	tenantUUID := tenants[0].GetId()
	s.T().Logf("Tenant Detail- Name: %s, UUID: %s ", tenantName, tenantUUID)
	s.Require().Equal("Default", tenantName, "Default tenant doesn't exists.")

	projects := ListProjects(s, tenantUUID)
	s.Require().Equal(1, len(projects))
	projectUUID := projects[0].GetId()
	projectName := projects[0].GetName()
	s.T().Logf("Project Detail- Name: %s, UUID: %s ", projectName, projectUUID)
	s.Require().Equal("Default", projectName, "Default project doesn't exists.")
}

func ListAccounts(s *PDSTestSuite) []pds.ModelsAccount {
	accountClient := s.apiClient.AccountsApi
	s.T().Log("Get list of Accounts.")
	accountsModel, res, err := accountClient.ApiAccountsGet(s.ctx).Execute()
	s.T().Log("API call succeeded.")
	require.NoError(s.T(), err)
	require.Equal(s.T(), status.StatusOK, res.StatusCode)
	return accountsModel.GetData()
}

func ListTenants(s *PDSTestSuite, uuid string) []pds.ModelsTenant {
	accountClient := s.apiClient.AccountsApi
	account, res, err := accountClient.ApiAccountsIdGet(s.ctx, uuid).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		panic("Unable to fetch the Accounts details.")
	}
	s.T().Logf("Get the list of Tenants belong to the account : %s", account.GetName())
	tenantClient := s.apiClient.TenantsApi
	tenantModels, res, err := tenantClient.ApiAccountsIdTenantsGet(s.ctx, account.GetId()).Execute()
	require.NoError(s.T(), err)
	require.Equal(s.T(), status.StatusOK, res.StatusCode)
	return tenantModels.GetData()
}

func ListProjects(s *PDSTestSuite, uuid string) []pds.ModelsProject {
	tenantClient := s.apiClient.TenantsApi
	tenant, res, err := tenantClient.ApiTenantsIdGet(s.ctx, uuid).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		panic("Unable to fetch the Tenant details.")
	}
	s.T().Logf("Get the list of projects belong to the tenant : %s", tenant.GetName())
	projectClient := s.apiClient.ProjectsApi
	projectModels, res, err := projectClient.ApiTenantsIdProjectsGet(s.ctx, tenant.GetId()).Execute()
	require.NoError(s.T(), err)
	require.Equal(s.T(), status.StatusOK, res.StatusCode)
	return projectModels.GetData()
}
