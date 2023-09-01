package iam_test

import (
	"context"
	"net/http"
	"time"

	"k8s.io/utils/pointer"

	pdsApi "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/random"
)

// TestUserAPIKey_SanityCheck tests the sanity of user api keys.
// Steps:
// 1. Create User API Key by invoking PDS Create UserAPI key
// 2. Using the token from (1) list accounts
// 3. Disable the token by invoking PDS Patch UserAPIKey
// 4. Using the token from (1) list accounts
// Expected:
// 1. User API key should be created
// 2. List accounts should return success
// 3. Disabling the API key should be success
// 4. List accounts should return error
func (s *IAMTestSuite) TestUserAPIKey_SanityCheck() {
	// Create a user api key.
	keyName := "test-api-key-" + random.AlphaNumericString(10)
	expireDate := time.Now().AddDate(0, 0, 1)
	key, err := s.ControlPlane.PDS.CreateUserAPIKey(expireDate, keyName)
	s.Require().NoError(err, "could not create api key")
	apiKeyClient, err := api.NewPDSClient(
		context.Background(),
		s.ControlPlane.PDS.URL,
		api.LoginCredentials{BearerToken: *key.JwtToken},
	)
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
