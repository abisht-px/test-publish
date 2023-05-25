package controlplane

import (
	"context"
	"net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) CreateInvitation(ctx context.Context, t tests.T, email, role string) (*http.Response, error) {
	requestBody := pds.RequestsInvitationAccountRequest{
		Email:    email,
		RoleName: role,
	}
	return c.PDS.AccountRoleBindingsApi.ApiAccountsIdInvitationsPost(ctx, c.testPDSAccountID).Body(requestBody).Execute()
}

func (c *ControlPlane) MustListAccountInvitations(ctx context.Context, t tests.T) *pds.ModelsPaginatedResultModelsAccountRoleInvitation {
	req := c.PDS.AccountsApi.ApiAccountsIdAccountRoleInvitationsGet(ctx, c.testPDSAccountID)
	invitations, resp, err := c.PDS.AccountsApi.ApiAccountsIdAccountRoleInvitationsGetExecute(req)
	api.RequireNoError(t, resp, err)
	require.NotEmpty(t, invitations.GetData())
	return invitations
}

func (c *ControlPlane) MustGetAccountInvitation(ctx context.Context, t tests.T, invitationID string) *pds.ModelsAccountRoleInvitation {
	req := c.PDS.AccountsApi.ApiAccountsIdAccountRoleInvitationsGet(ctx, c.testPDSAccountID)
	req.Id2(invitationID)
	invitations, resp, err := c.PDS.AccountsApi.ApiAccountsIdAccountRoleInvitationsGetExecute(req)
	api.RequireNoError(t, resp, err)
	require.NotEmpty(t, invitations.GetData())
	return &invitations.GetData()[0]
}

func (c *ControlPlane) MustPatchAccountInvitation(ctx context.Context, t tests.T, role, invitationID string) *http.Response {
	req := c.PDS.AccountsRoleInvitationsApi.ApiAccountRoleInvitationsIdPatch(ctx, invitationID)
	patchReqBody := pds.RequestsPatchAccountRoleInvitationRequest{RoleName: &role}
	req = req.Body(patchReqBody)
	resp, err := c.PDS.AccountsRoleInvitationsApi.ApiAccountRoleInvitationsIdPatchExecute(req)
	api.RequireNoError(t, resp, err)
	return resp
}

func (c *ControlPlane) MustDeleteInvitation(ctx context.Context, t tests.T, id string) *http.Response {
	req := c.PDS.AccountsRoleInvitationsApi.ApiAccountRoleInvitationsIdDelete(ctx, id)
	resp, err := c.PDS.AccountsRoleInvitationsApi.ApiAccountRoleInvitationsIdDeleteExecute(req)
	api.RequireNoError(t, resp, err)
	return resp
}
