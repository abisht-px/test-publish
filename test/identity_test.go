package test

import (
	"context"
	"net/http"

	"github.com/portworx/pds-integration-test/internal/api"
)

func (s *ControlPlaneTestSuite) TestIdentityMangement_SanityCheck() {

	authUserClient, _ := api.NewPDSClient(context.Background(), s.ControlPlane.PDS.URL, s.config.LoginCredentialsAuthUser)

	// Get the info about the auth user.
	authUserResponse, response, err := authUserClient.WhoAmIApi.ApiWhoamiGet(context.Background()).Execute()
	api.RequireNoError(s.T(), response, err)
	user, userOk := authUserResponse.GetUserOk()
	s.Require().True(userOk)
	s.Require().NotNil(user.GetId())

	// Ensure that user has no role in the account at the end.
	s.T().Cleanup(func() {
		s.ControlPlane.MustDeleteUserAccountRole(s.ctx, s.T(), user.GetId())
	})

	// Ensure the auth user has admin role in the account.
	s.ControlPlane.MustEnsureUserAccountRole(s.ctx, s.T(), user.GetId(), "account-admin")

	serviceAccounts, response, err := authUserClient.ServiceAccountsApi.ApiTenantsIdServiceAccountsGet(context.Background(), s.ControlPlane.TestPDSTenantID).Execute()
	api.RequireNoErrorWithStatus(s.T(), response, err, http.StatusOK)
	s.Require().True(serviceAccounts.HasData())

	// Ensure the auth user has reader role in the account.
	s.ControlPlane.MustEnsureUserAccountRole(s.ctx, s.T(), user.GetId(), "account-reader")

	// List the service accounts.
	result, response, err := authUserClient.ServiceAccountsApi.ApiTenantsIdServiceAccountsGet(context.Background(), s.ControlPlane.TestPDSTenantID).Execute()
	api.RequireErrorWithStatus(s.T(), response, err, http.StatusForbidden)
	s.Require().Nil(result)

	// List the deployment targets.
	clusters, response, err := authUserClient.DeploymentTargetsApi.ApiTenantsIdDeploymentTargetsGet(context.Background(), s.ControlPlane.TestPDSTenantID).Execute()
	api.RequireNoErrorWithStatus(s.T(), response, err, http.StatusOK)
	s.Require().True(clusters.HasData())

	// Ensure the auth user has no role in the account.
	s.ControlPlane.MustDeleteUserAccountRole(s.ctx, s.T(), user.GetId())

	// List the deployment targets.
	targets, response, err := authUserClient.DeploymentTargetsApi.ApiTenantsIdDeploymentTargetsGet(context.Background(), s.ControlPlane.TestPDSTenantID).Execute()
	api.RequireErrorWithStatus(s.T(), response, err, http.StatusForbidden)
	s.Require().Nil(targets)

}
