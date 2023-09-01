package controlplane

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/require"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) MustCreateIAM(ctx context.Context,
	t tests.T, actorID string, policy pds.ModelsAccessPolicy) *pds.ModelsIAM {
	iam, resp, err := c.CreateIAM(ctx, actorID, policy)
	api.RequireNoError(t, resp, err)
	require.NotNil(t, iam)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return iam
}

func (c *ControlPlane) CreateIAM(ctx context.Context,
	actorID string, policy pds.ModelsAccessPolicy) (*pds.ModelsIAM, *http.Response, error) {

	r := c.PDS.IAMApi.ApiAccountsIdIamPost(ctx, c.TestPDSAccountID)
	r = r.Body(*pds.NewRequestsIAMRequest(actorID, policy))
	return c.PDS.IAMApi.ApiAccountsIdIamPostExecute(r)
}

func (c *ControlPlane) MustUpdateIAM(ctx context.Context, t tests.T, actorID string,
	accessPolicy pds.ModelsAccessPolicy) *pds.ModelsIAM {
	iam, resp, err := c.UpdateIAM(ctx, actorID, accessPolicy)
	api.RequireNoError(t, resp, err)
	require.NotNil(t, iam)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return iam
}

func (c *ControlPlane) UpdateIAM(ctx context.Context, actorID string,
	accessPolicy pds.ModelsAccessPolicy) (*pds.ModelsIAM, *http.Response, error) {

	r := c.PDS.IAMApi.ApiAccountsIdIamPut(ctx, c.TestPDSAccountID)
	r = r.Body(pds.RequestsIAMRequest{ActorId: actorID, Data: accessPolicy})
	return c.PDS.IAMApi.ApiAccountsIdIamPutExecute(r)
}

func (c *ControlPlane) MustDeleteIAM(ctx context.Context, t tests.T, actorID string) {
	r := c.PDS.IAMApi.ApiAccountsIdIamActorIdDelete(ctx, c.TestPDSAccountID, actorID)
	resp, err := c.PDS.IAMApi.ApiAccountsIdIamActorIdDeleteExecute(r)
	api.RequireNoError(t, resp, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func (c *ControlPlane) ListIAM(ctx context.Context, t tests.T) ([]pds.ModelsIAM, *http.Response, error) {
	r := c.PDS.IAMApi.ApiAccountsIdIamGet(ctx, c.TestPDSAccountID)
	return c.PDS.IAMApi.ApiAccountsIdIamGetExecute(r)
}

func (c *ControlPlane) GetIAM(ctx context.Context, t tests.T, actorID string) (*pds.ModelsIAM, *http.Response, error) {
	r := c.PDS.IAMApi.ApiAccountsIdIamActorIdGet(ctx, c.TestPDSAccountID, actorID)
	return c.PDS.IAMApi.ApiAccountsIdIamActorIdGetExecute(r)
}
