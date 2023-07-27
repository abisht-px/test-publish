package controlplane

import (
	"context"
	"net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/assert"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

var actorType string = "user"

func (c *ControlPlane) mustSetRoleBinding(ctx context.Context, t tests.T, userID string, userRole string) {

	accountBinding, response, err := c.PDS.AccountRoleBindingsApi.
		ApiAccountsIdRoleBindingsPut(ctx, c.TestPDSAccountID).
		Body(pds.RequestsPutLegacyBindingRequest{ActorId: &userID, ActorType: &actorType, RoleName: &userRole}).
		Execute()

	api.RequireNoError(t, response, err)
	assert.Equal(t, http.StatusOK, response.StatusCode, "Error updating rolebinding for userID %s", userID)
	assert.Equal(t, userRole, accountBinding.GetRoleName(), "Error updating rolebinding for userID %s", userID)
}

func (c *ControlPlane) MustEnsureUserAccountRole(ctx context.Context, t tests.T, userID string, userRole string) {

	accRoleBindings, response, err := c.PDS.AccountRoleBindingsApi.
		ApiAccountsIdRoleBindingsGet(ctx, c.TestPDSAccountID).
		ActorId(userID).
		Execute()

	api.RequireNoError(t, response, err)
	roleBinding, ok := accRoleBindings.GetDataOk()
	assert.Truef(t, ok, "Error getting rolebinding for userID %s", userID)

	assert.Lessf(t, len(roleBinding), 2, "Found multiple rolebindings for active account and userID %s", userID)

	if len(roleBinding) == 0 || userRole != roleBinding[0].GetRoleName() {
		c.mustSetRoleBinding(ctx, t, userID, userRole)
	}
}

func (c *ControlPlane) MustDeleteUserAccountRole(ctx context.Context, t tests.T, userID string) {
	response, err := c.PDS.AccountRoleBindingsApi.
		ApiAccountsIdRoleBindingsDelete(ctx, c.TestPDSAccountID).
		Body(pds.RequestsDeleteRoleBindingRequest{ActorId: &userID, ActorType: &actorType}).
		Execute()
	if response != nil && response.StatusCode == http.StatusNotFound {
		return
	}
	api.RequireNoError(t, response, err)
	assert.Equal(t, http.StatusNoContent, response.StatusCode, "Error deleting rolebinding for userID %s", userID)
}
