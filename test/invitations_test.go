package test

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/portworx/pds-integration-test/internal/api"
)

const (
	testInvitationEmail = "test_invitation_integration_test@email.com"
)

func (s *PDSTestSuite) TestInvitations_CrudFail() {
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

func (s *PDSTestSuite) TestInvitations_CrudOK() {
	// create invitation.
	response, err := s.controlPlane.CreateInvitation(s.ctx, s.T(), testInvitationEmail, "account-reader")
	api.RequireNoError(s.T(), response, err)
	s.Require().Equal(http.StatusOK, response.StatusCode)

	// Get invitation.
	createdInvitation := s.controlPlane.GetAccountInvitation(s.ctx, s.T(), testInvitationEmail)
	s.Require().NotNil(createdInvitation)

	// checking patch invitation.
	s.controlPlane.MustPatchAccountInvitation(s.ctx, s.T(), "account-admin", *createdInvitation.Id)

	fetchedInvitation := s.controlPlane.GetAccountInvitation(s.ctx, s.T(), *createdInvitation.Email)
	// verifying patch.
	s.Require().Equal("account-admin", fetchedInvitation.GetRoleName())

	// deleting invitation.
	s.controlPlane.MustDeleteInvitation(s.ctx, s.T(), *createdInvitation.Id)

	// verifying deletion.
	deletedInvitations := s.controlPlane.GetAccountInvitation(s.ctx, s.T(), *createdInvitation.Email)
	s.Require().Nil(deletedInvitations)
}
