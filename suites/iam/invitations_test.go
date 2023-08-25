package iam_test

import (
	"net/http"

	"github.com/google/uuid"
)

const (
	testInvitationEmail = "test_invitation_integration_test@email.com"
)

func (s *IAMTestSuite) TestInvitations_CreateFail() {
	// invalid email.
	response, err := s.ControlPlane.CreateInvitation(s.ctx, s.T(), "invalid-email-id", "account-reader")
	s.Require().Error(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusBadRequest, response.StatusCode)

	// valid email and invalid user role.
	response, err = s.ControlPlane.CreateInvitation(s.ctx, s.T(), testInvitationEmail, "invalid-user-role")
	s.Require().Error(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusUnprocessableEntity, response.StatusCode)

	s.ControlPlane.MustCreateInvitation(s.ctx, s.T(), testInvitationEmail, "account-reader")

	response, err = s.ControlPlane.CreateInvitation(s.ctx, s.T(), testInvitationEmail, "account-reader")
	s.Require().Error(err)
	s.Require().Equal(http.StatusConflict, response.StatusCode)

	fetchedInvitation := s.ControlPlane.GetAccountInvitation(s.ctx, s.T(), testInvitationEmail)
	s.Require().NotNil(fetchedInvitation)
	s.ControlPlane.MustDeleteInvitation(s.ctx, s.T(), *fetchedInvitation.Id)
}

func (s *IAMTestSuite) TestInvitations_DeleteFail() {
	// deleting invitation with id does not exists.
	response, err := s.ControlPlane.DeleteInvitation(s.ctx, uuid.New().String())
	s.Require().Error(err)
	s.Require().Equal(http.StatusNotFound, response.StatusCode)

	// deleting invitation with invalid invitation id.
	response, err = s.ControlPlane.DeleteInvitation(s.ctx, "invalid_invitation_id")
	s.Require().Error(err)
	s.Require().Equal(http.StatusBadRequest, response.StatusCode)
}

func (s *IAMTestSuite) TestInvitations_CrudOK() {
	s.T().Skip("Need to fix the test")

	// create invitation.
	s.ControlPlane.MustCreateInvitation(s.ctx, s.T(), testInvitationEmail, "account-reader")

	// Get invitation.
	createdInvitation := s.ControlPlane.GetAccountInvitation(s.ctx, s.T(), testInvitationEmail)
	s.Require().NotNil(createdInvitation)

	// checking patch invitation.
	s.ControlPlane.MustPatchAccountInvitation(s.ctx, s.T(), "account-admin", *createdInvitation.Id)

	fetchedInvitation := s.ControlPlane.GetAccountInvitation(s.ctx, s.T(), *createdInvitation.Email)
	// verifying patch.
	s.Require().Equal("account-admin", fetchedInvitation.GetRoleName())

	// deleting invitation.
	s.ControlPlane.MustDeleteInvitation(s.ctx, s.T(), *createdInvitation.Id)

	// verifying deletion.
	deletedInvitations := s.ControlPlane.GetAccountInvitation(s.ctx, s.T(), *createdInvitation.Email)
	s.Require().Nil(deletedInvitations)
}
