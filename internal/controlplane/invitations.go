package controlplane

import (
	"context"
	"net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) MustCreateInvitation(ctx context.Context, t tests.T, email, accID, role string) *http.Response {
	resp, err := c.CreateInvitation(ctx, email, accID, role)
	api.RequireNoError(t, resp, err)
	return resp
}

func (c *ControlPlane) CreateInvitation(ctx context.Context, email, id, role string) (*http.Response, error) {
	requestBody := pds.RequestsInvitationAccountRequest{
		Email:    email,
		RoleName: role,
	}
	return c.PDS.AccountRoleBindingsApi.ApiAccountsIdInvitationsPost(ctx, id).Body(requestBody).Execute()
}

func (c *ControlPlane) MustListAccountInvitations(ctx context.Context, t tests.T, id string) *pds.ModelsPaginatedResultModelsAccountRoleInvitation {
	req := c.PDS.AccountsApi.ApiAccountsIdAccountRoleInvitationsGet(ctx, id)
	invitations, resp, err := c.PDS.AccountsApi.ApiAccountsIdAccountRoleInvitationsGetExecute(req)
	api.RequireNoError(t, resp, err)
	require.NotEmpty(t, invitations)
	return invitations
}

func (c *ControlPlane) MustGetAccountInvitation(ctx context.Context, t tests.T, accID, invitationID string) *pds.ModelsAccountRoleInvitation {
	req := c.PDS.AccountsApi.ApiAccountsIdAccountRoleInvitationsGet(ctx, accID)
	req.Id2(invitationID)
	invitation, resp, err := c.PDS.AccountsApi.ApiAccountsIdAccountRoleInvitationsGetExecute(req)
	api.RequireNoError(t, resp, err)
	require.NotNil(t, invitation)
	return &invitation.GetData()[0]
}

func (c *ControlPlane) MustDeleteInvitation(ctx context.Context, t tests.T, id string) {
	req := c.PDS.AccountsRoleInvitationsApi.ApiAccountRoleInvitationsIdDelete(ctx, id)
	resp, err := c.PDS.AccountsRoleInvitationsApi.ApiAccountRoleInvitationsIdDeleteExecute(req)
	api.RequireNoError(t, resp, err)
}
