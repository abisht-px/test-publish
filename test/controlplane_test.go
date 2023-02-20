package test

import (
	"net/http"
	"time"

	"github.com/portworx/pds-integration-test/internal/random"
	"github.com/portworx/pds-integration-test/test/api"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"golang.org/x/net/context"
	"k8s.io/utils/pointer"
)

func (s *PDSTestSuite) TestUserAPIKey_SanityCheck() {
	// Create a user api key.
	keyName := "test-api-key-" + random.AlphaNumericString(10)
	expireDate := time.Date(2050, time.January, 1, 1, 1, 1, 1, time.UTC)
	key := s.mustCreateUserAPIKey(s.ctx, s.apiClient, expireDate, keyName)
	// Create a context with this key for the apiClient.
	userApiKeyContext := context.WithValue(s.ctx, pds.ContextAPIKeys,
		map[string]pds.APIKey{
			"ApiKeyAuth": {Key: *key.JwtToken, Prefix: "Bearer"},
		},
	)

	// Try the token: list accounts.
	paginatedResult, response, err := s.apiClient.AccountsApi.ApiAccountsGet(userApiKeyContext).Execute()
	accounts := paginatedResult.GetData()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.Require().NotEmpty(accounts)

	// Disable the token.
	response, err = s.apiClient.UserAPIKeyApi.ApiUserApiKeyIdPatch(s.ctx, *key.Id).Body(
		pds.RequestsPatchUserAPIKeyRequest{
			Enabled: pointer.Bool(false),
		}).Execute()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)

	// Try the disabled token: fail on list accounts.
	_, response, _ = s.apiClient.AccountsApi.ApiAccountsGet(userApiKeyContext).Execute()
	s.Require().Equal(http.StatusUnauthorized, response.StatusCode)

	// Delete the token.
	response, err = s.apiClient.UserAPIKeyApi.ApiUserApiKeyIdDelete(s.ctx, *key.Id).Execute()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusNoContent, response.StatusCode)

	// Try the deleted token: fail on list accounts.
	_, response, _ = s.apiClient.AccountsApi.ApiAccountsGet(userApiKeyContext).Execute()
	s.Require().Equal(http.StatusForbidden, response.StatusCode)
}
