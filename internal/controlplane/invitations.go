package controlplane

import (
	"context"
	"net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) MustCreateInvitation(ctx context.Context, t tests.T, email, role string) {
	response, err := c.CreateInvitation(ctx, t, email, role)
	api.RequireNoError(t, response, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func (c *ControlPlane) CreateInvitation(ctx context.Context, t tests.T, email, role string) (*http.Response, error) {
	requestBody := pds.RequestsInvitationAccountRequest{
		Email:    email,
		RoleName: role,
	}
	return c.PDS.AccountRoleBindingsApi.ApiAccountsIdInvitationsPost(ctx, c.TestPDSAccountID).Body(requestBody).Execute()
}

func (c *ControlPlane) MustListAccountInvitations(ctx context.Context, t tests.T) *pds.ModelsPaginatedResultModelsAccountRoleInvitation {
	req := c.PDS.AccountsApi.ApiAccountsIdAccountRoleInvitationsGet(ctx, c.TestPDSAccountID)
	invitations, resp, err := c.PDS.AccountsApi.ApiAccountsIdAccountRoleInvitationsGetExecute(req)
	require.NotNil(t, invitations)
	api.RequireNoError(t, resp, err)
	require.NotEmpty(t, invitations.GetData())
	return invitations
}

func (c *ControlPlane) GetAccountInvitation(ctx context.Context, t tests.T, email string) *pds.ModelsAccountRoleInvitation {
	invitations := c.MustListAccountInvitations(ctx, t)
	require.NotEmpty(t, invitations.GetData())
	for _, invitation := range invitations.GetData() {
		if *invitation.Email == email {
			return &invitation
		}
	}
	return nil
}

func (c *ControlPlane) MustPatchAccountInvitation(ctx context.Context, t tests.T, role, invitationID string) {
	req := c.PDS.AccountsRoleInvitationsApi.ApiAccountRoleInvitationsIdPatch(ctx, invitationID)
	patchReqBody := pds.RequestsPatchAccountRoleInvitationRequest{RoleName: &role}
	req = req.Body(patchReqBody)
	resp, err := c.PDS.AccountsRoleInvitationsApi.ApiAccountRoleInvitationsIdPatchExecute(req)
	api.RequireNoError(t, resp, err)
}

func (c *ControlPlane) MustDeleteInvitation(ctx context.Context, t tests.T, id string) {
	resp, err := c.DeleteInvitation(ctx, id)
	require.NotNil(t, resp)
	api.RequireNoError(t, resp, err)
}

func (c *ControlPlane) DeleteInvitation(ctx context.Context, id string) (*http.Response, error) {
	return c.PDS.AccountsRoleInvitationsApi.ApiAccountRoleInvitationsIdDelete(ctx, id).Execute()
}
