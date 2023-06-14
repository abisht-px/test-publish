package test

import (
	"net/http"

	"github.com/google/uuid"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
)

const (
	testInvitationEmail = "test_invitation_integration_test@email.com"
)

func (s *PDSTestSuite) TestInvitations_CreateInvitation_Fail() {
	// invalid email.
	response, err := s.controlPlane.CreateInvitation(s.ctx, s.T(), "invalid-email-id", "account-reader")
	s.Require().Error(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusBadRequest, response.StatusCode)

	// valid email and invalid user role.
	response, err = s.controlPlane.CreateInvitation(s.ctx, s.T(), testInvitationEmail, "invalid-user-role")
	s.Require().Error(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusUnprocessableEntity, response.StatusCode)

	// deleting invitation with id does not exists.
	response, err = s.controlPlane.DeleteInvitation(s.ctx, uuid.New().String())
	s.Require().Error(err)
	s.Require().Equal(http.StatusNotFound, response.StatusCode)

	// deleting invitation with invalid invitation id.
	response, err = s.controlPlane.DeleteInvitation(s.ctx, "invalid_invitation_id")
	s.Require().Error(err)
	s.Require().Equal(http.StatusBadRequest, response.StatusCode)
}

func (s *PDSTestSuite) TestInvitations_CreateInvitation_CRUD_OK() {
	// create invitation.
	response, err := s.controlPlane.CreateInvitation(s.ctx, s.T(), testInvitationEmail, "account-reader")
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)

	// List invitations.
	result := s.controlPlane.MustListAccountInvitations(s.ctx, s.T())
	found := false
	var createdInvitation pds.ModelsAccountRoleInvitation
	for _, invitation := range result.Data {
		if *invitation.Email == testInvitationEmail {
			found = true
			createdInvitation = invitation
			break
		}
	}
	s.Require().True(found)

	// checking patch invitation.
	response = s.controlPlane.MustPatchAccountInvitation(s.ctx, s.T(), "account-admin", *createdInvitation.Id)
	s.Require().Equal(http.StatusNoContent, response.StatusCode)

	fetchedInvitation := s.controlPlane.GetAccountInvitation(s.ctx, s.T(), *createdInvitation.Id)
	// verifying patch.
	s.Require().Equal("account-admin", fetchedInvitation.GetRoleName())

	// deleting invitation.
	response = s.controlPlane.MustDeleteInvitation(s.ctx, s.T(), *createdInvitation.Id)
	s.Require().Equal(http.StatusNoContent, response.StatusCode)

	// verifying deletion.
	deletedInvitations := s.controlPlane.GetAccountInvitation(s.ctx, s.T(), *createdInvitation.Id)
	s.Require().Nil(deletedInvitations)
}
