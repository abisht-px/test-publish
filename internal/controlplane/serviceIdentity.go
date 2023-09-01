package controlplane

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/require"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) MustCreateServiceIdentity(ctx context.Context, t tests.T, accountID, name string, enabled bool) *pds.ModelsServiceIdentityWithToken {
	token, resp, err := c.CreateServiceIdentity(ctx, accountID, name, enabled)
	api.RequireNoError(t, resp, err)
	require.NotNil(t, token)
	return token
}

func (c *ControlPlane) CreateServiceIdentity(ctx context.Context, accountID, name string, enabled bool) (*pds.ModelsServiceIdentityWithToken, *http.Response, error) {
	requestBody := pds.RequestsServiceIdentityRequest{
		Name:    name,
		Enabled: enabled,
	}

	return c.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityPost(ctx, accountID).Body(requestBody).Execute()
}

func (c *ControlPlane) GetServiceIdentity(ctx context.Context, t tests.T, serviceIdentityID string) (*pds.ModelsServiceIdentity, *http.Response, error) {
	return c.PDS.ServiceIdentityApi.ApiServiceIdentityIdGet(ctx, serviceIdentityID).Execute()
}

func (c *ControlPlane) MustUpdateServiceIdentity(ctx context.Context, t tests.T, serviceIdentityID string, requestBody *pds.RequestsServiceIdentityRequest) {
	resp, err := c.UpdateServiceIdentity(ctx, serviceIdentityID, requestBody)
	api.RequireNoError(t, resp, err)
}

func (c *ControlPlane) UpdateServiceIdentity(ctx context.Context, serviceIdentityID string, requestBody *pds.RequestsServiceIdentityRequest) (*http.Response, error) {
	return c.PDS.ServiceIdentityApi.ApiServiceIdentityIdPut(ctx, serviceIdentityID).Body(*requestBody).Execute()
}

func (c *ControlPlane) DeleteServiceIdentity(ctx context.Context, serviceIdentityID string) (*http.Response, error) {
	return c.PDS.ServiceIdentityApi.ApiServiceIdentityIdDelete(ctx, serviceIdentityID).Execute()
}

func (c *ControlPlane) RegenerateServiceIdentity(ctx context.Context, serviceIdentityID string) (*pds.ModelsServiceIdentityWithToken, *http.Response, error) {
	return c.PDS.ServiceIdentityApi.ApiServiceIdentityIdRegenerateGet(ctx, serviceIdentityID).Execute()
}

func (c *ControlPlane) GenerateTokenServiceIdentity(ctx context.Context, requestBody *pds.ControllersGenerateTokenRequest) (*pds.ControllersGenerateTokenResponse, *http.Response, error) {
	return c.PDS.ServiceIdentityApi.ServiceIdentityGenerateTokenPost(ctx).Body(*requestBody).Execute()
}

func (c *ControlPlane) ListServiceIdentity(ctx context.Context, t tests.T, accountID string) (*[]pds.ModelsServiceIdentity, *http.Response, error) {

	serviceIdentity, resp, err := c.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(ctx, accountID).SortBy("created_by").Execute()
	if err != nil {
		return nil, resp, err
	}
	api.RequireNoError(t, resp, err)
	require.NotEmpty(t, serviceIdentity)
	return &serviceIdentity.Data, resp, err
}
