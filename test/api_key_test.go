package test

import (
	"context"
	"net/http"
	"time"

	"github.com/portworx/pds-integration-test/internal/random"

	"github.com/portworx/pds-integration-test/internal/api"

	pdsApi "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"k8s.io/utils/pointer"
)

func (s *ControlPlaneTestSuite) TestUserAPIKey_SanityCheck() {
	// Create a user api key.
	keyName := "test-api-key-" + random.AlphaNumericString(10)
	expireDate := time.Now().AddDate(0, 0, 1)
	key, err := s.ControlPlane.PDS.CreateUserAPIKey(expireDate, keyName)
	s.Require().NoError(err, "could not create api key")
	apiKeyClient, err := api.NewPDSClient(context.Background(), s.ControlPlane.PDS.URL, api.LoginCredentials{BearerToken: *key.JwtToken})
	s.Require().NoError(err, "could not create auth context for api call")

	// Try the token: list accounts.
	paginatedResult, response, err := apiKeyClient.AccountsApi.ApiAccountsGet(context.Background()).Execute()
	accounts := paginatedResult.GetData()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.Require().NotEmpty(accounts)

	// Disable the token.
	response, err = s.ControlPlane.PDS.UserAPIKeyApi.ApiUserApiKeyIdPatch(context.Background(), *key.Id).Body(
		pdsApi.RequestsPatchUserAPIKeyRequest{
			Enabled: pointer.Bool(false),
		}).Execute()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)

	// Try the disabled token: fail on list accounts.
	_, response, _ = apiKeyClient.AccountsApi.ApiAccountsGet(context.Background()).Execute()
	s.Require().Equal(http.StatusUnauthorized, response.StatusCode)

	// Enable the token again.
	response, err = s.ControlPlane.PDS.UserAPIKeyApi.ApiUserApiKeyIdPatch(context.Background(), *key.Id).Body(
		pdsApi.RequestsPatchUserAPIKeyRequest{
			Enabled: pointer.Bool(true),
		}).Execute()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)

	// Try the token: list accounts.
	paginatedResult, response, err = apiKeyClient.AccountsApi.ApiAccountsGet(context.Background()).Execute()
	accounts = paginatedResult.GetData()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.Require().NotEmpty(accounts)

	// Disable the token.
	response, err = s.ControlPlane.PDS.UserAPIKeyApi.ApiUserApiKeyIdPatch(context.Background(), *key.Id).Body(
		pdsApi.RequestsPatchUserAPIKeyRequest{
			Enabled: pointer.Bool(false),
		}).Execute()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)

	// Delete the token.
	response, err = s.ControlPlane.PDS.UserAPIKeyApi.ApiUserApiKeyIdDelete(context.Background(), *key.Id).Execute()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusNoContent, response.StatusCode)

	// Try the deleted token: fail on list accounts.
	_, response, _ = apiKeyClient.AccountsApi.ApiAccountsGet(context.Background()).Execute()
	s.Require().Equal(http.StatusForbidden, response.StatusCode)
}
