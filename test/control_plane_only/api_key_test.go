package control_plane_only

import (
	"context"
	"net/http"
	"time"

	"github.com/portworx/pds-integration-test/internal/random"

	"github.com/portworx/pds-integration-test/internal/api"

	pdsApi "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	pds "github.com/portworx/pds-integration-test/internal/pds"

	"k8s.io/utils/pointer"
)

func (s *ControlPlaneTestSuite) TestUserAPIKey_SanityCheck() {
	// Create a user api key.
	keyName := "test-api-key-" + random.AlphaNumericString(10)
	expireDate := time.Date(2025, time.January, 1, 1, 1, 1, 1, time.UTC)
	key, err := s.ControlPlane.CreateUserAPIKey(expireDate, keyName, s.ControlPlane.GetDefaultActor())
	s.Require().NoError(err, "could not create api key")
	apiKeyAuthCtx := context.Background()
	apiKeyAuthCtx, err = pds.CreateAuthContext(
		apiKeyAuthCtx, pds.LoginCredentials{BearerToken: *key.JwtToken})
	s.Require().NoError(err, "could not create auth context for api call")
	defaultUser := s.ControlPlane.GetDefaultActor()

	// Try the token: list accounts.
	paginatedResult, response, err := s.ControlPlane.ApiClient.AccountsApi.ApiAccountsGet(apiKeyAuthCtx).Execute()
	accounts := paginatedResult.GetData()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.Require().NotEmpty(accounts)

	// Disable the token.
	response, err = s.ControlPlane.ApiClient.UserAPIKeyApi.ApiUserApiKeyIdPatch(defaultUser.AuthCtx, *key.Id).Body(
		pdsApi.RequestsPatchUserAPIKeyRequest{
			Enabled: pointer.Bool(false),
		}).Execute()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)

	// Try the disabled token: fail on list accounts.
	_, response, _ = s.ControlPlane.ApiClient.AccountsApi.ApiAccountsGet(apiKeyAuthCtx).Execute()
	s.Require().Equal(http.StatusUnauthorized, response.StatusCode)

	// Delete the token.
	response, err = s.ControlPlane.ApiClient.UserAPIKeyApi.ApiUserApiKeyIdDelete(defaultUser.AuthCtx, *key.Id).Execute()
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusNoContent, response.StatusCode)

	// Try the deleted token: fail on list accounts.
	_, response, _ = s.ControlPlane.ApiClient.AccountsApi.ApiAccountsGet(apiKeyAuthCtx).Execute()
	s.Require().Equal(http.StatusForbidden, response.StatusCode)
}
