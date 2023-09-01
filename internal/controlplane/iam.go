package controlplane

import (
	"context"
	"net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) MustCreateIAM(ctx context.Context,
	t tests.T, roleName, actorID string) *pds.ModelsIAM {
	policy := pds.ModelsAccessPolicy{
		Account: []string{roleName},
	}
	iam, resp, err := c.CreateIAM(ctx, actorID, policy)
	api.RequireNoError(t, resp, err)
	return iam
}

func (c *ControlPlane) CreateIAM(ctx context.Context,
	actorID string, policy pds.ModelsAccessPolicy) (*pds.ModelsIAM, *http.Response, error) {

	r := c.PDS.IAMApi.ApiAccountsIdIamPost(ctx, c.TestPDSAccountID)
	r = r.Body(*pds.NewRequestsIAMRequest(actorID, policy))
	return c.PDS.IAMApi.ApiAccountsIdIamPostExecute(r)
}

func (c *ControlPlane) GetIAM(ctx context.Context, t tests.T, actorID string) (*pds.ModelsIAM, *http.Response, error) {
	r := c.PDS.IAMApi.ApiAccountsIdIamActorIdGet(ctx, c.TestPDSAccountID, actorID)
	return c.PDS.IAMApi.ApiAccountsIdIamActorIdGetExecute(r)
}
