package controlplane

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) mustGetServiceAccountID(ctx context.Context, t tests.T, name string) string {
	// TODO: Use service account name query filters
	serviceAccounts, resp, err := c.API.ServiceAccountsApi.ApiTenantsIdServiceAccountsGet(ctx, c.TestPDSTenantID).Execute()
	api.RequireNoErrorf(t, resp, err, "Getting service account %s under tenant %s.", name, c.TestPDSTenantID)
	require.NotEmpty(t, serviceAccounts, "PDS API must return at least one tenant.")

	var serviceAccountID string
	for _, serviceAccount := range serviceAccounts.GetData() {
		if serviceAccount.GetName() == name {
			serviceAccountID = serviceAccount.GetId()
			break
		}
	}
	require.NotEmpty(t, serviceAccountID, "PDS service account %s not found.", name)
	return serviceAccountID
}

func (c *ControlPlane) MustGetServiceAccountToken(ctx context.Context, t tests.T, serviceAccountName string) string {
	serviceAccountID := c.mustGetServiceAccountID(ctx, t, serviceAccountName)
	token, resp, err := c.API.ServiceAccountsApi.ApiServiceAccountsIdTokenGet(ctx, serviceAccountID).Execute()
	api.RequireNoErrorf(t, resp, err, "Getting token for service account %s.", serviceAccountName)
	require.Equal(t, http.StatusOK, resp.StatusCode, "PDS API must return HTTP 200.")

	return token.GetToken()
}
