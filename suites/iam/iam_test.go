package iam_test

import (
	"net/http"
	"testing"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/suites/framework"
)

var (
	tenantAdmin   = "tenant-admin"
	accountAdmin  = "account-admin"
	accountReader = "account-reader"
	projectRole   = "project-admin"
)

func (s *IAMTestSuite) TestIAM() {
	account, err := s.ControlPlane.PDS.GetAccount(framework.PDSAccountName)
	s.Require().NoError(err)
	s.Require().NotNil(account)

	tenant, err := s.ControlPlane.PDS.GetTenant(*account.Id, framework.PDSTenantName)
	s.Require().NoError(err)
	s.Require().NotNil(tenant)

	// Fetching project details to set project level roles.
	project, err := s.ControlPlane.PDS.GetProject(*tenant.Id, framework.PDSProjectName)
	s.Require().NoError(err)
	s.Require().NotNil(project)

	s.testIAM_Create(*tenant.Id)
	s.testIAM_Update(*tenant.Id)
	s.testIAM_Update_UpdateByRemoving(*tenant.Id, *project.Id)
	s.testIAM_Update_GetList(*tenant.Id)
	s.testIAM_VerifyAuth(*tenant.Id, *project.Id)
}

func (s *IAMTestSuite) testIAM_Create(tenantID string) {
	testCases := []struct {
		testName      string
		policy        pds.ModelsAccessPolicy
		userID        string
		responseCode  int
		expectedError bool
		doCleanup     bool
	}{
		{
			testName: "create IAM -- invalid userID",
			policy: pds.ModelsAccessPolicy{
				Account: []string{accountAdmin},
			},
			userID:        "invalid-user",
			responseCode:  http.StatusBadRequest,
			expectedError: true,
			doCleanup:     false,
		},
		{
			testName:      "create IAM -- empty policy",
			policy:        pds.ModelsAccessPolicy{},
			userID:        testUserID,
			responseCode:  http.StatusUnprocessableEntity,
			expectedError: true,
			doCleanup:     false,
		},
		{
			testName: "create IAM -- invalid tenant",
			policy: pds.ModelsAccessPolicy{
				Tenant: []pds.ModelsBinding{
					{
						RoleName:    &tenantAdmin,
						ResourceIds: []string{"invalid-tenant"},
					},
				},
			},
			userID:        testUserID,
			responseCode:  http.StatusBadRequest,
			expectedError: true,
			doCleanup:     false,
		},
		{
			testName: "create IAM -- invalid tenant-role",
			policy: pds.ModelsAccessPolicy{
				Tenant: []pds.ModelsBinding{
					{
						RoleName:    &accountAdmin,
						ResourceIds: []string{tenantID},
					},
				},
			},
			userID:        testUserID,
			responseCode:  http.StatusUnprocessableEntity,
			expectedError: true,
			doCleanup:     false,
		},
		{
			testName: "create IAM -- create should be successful",
			policy: pds.ModelsAccessPolicy{
				Tenant: []pds.ModelsBinding{
					{
						RoleName:    &tenantAdmin,
						ResourceIds: []string{tenantID},
					},
				},
			},
			userID:        testUserID,
			responseCode:  http.StatusOK,
			expectedError: false,
			doCleanup:     true,
		},
		{
			testName: "create IAM -- account role create should be successful",
			policy: pds.ModelsAccessPolicy{
				Account: []string{accountAdmin},
			},
			userID:        testUserID,
			responseCode:  http.StatusOK,
			expectedError: false,
			doCleanup:     true,
		},
		{
			testName: "create IAM -- global role create should be successful",
			policy: pds.ModelsAccessPolicy{
				Global: []string{"pds-base"},
			},
			userID:        testUserID,
			responseCode:  http.StatusOK,
			expectedError: false,
			doCleanup:     true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.testName, func(t *testing.T) {
			_, response, err := s.ControlPlane.CreateIAM(s.ctx, tc.userID, tc.policy)
			s.checkError(err, tc.expectedError)
			s.Require().Equal(tc.responseCode, response.StatusCode)
			if tc.doCleanup {
				s.ControlPlane.MustDeleteIAM(s.ctx, s.T(), testUserID)
			}
		})
	}
}

func (s *IAMTestSuite) testIAM_Update(tenantID string) {
	policy := pds.ModelsAccessPolicy{
		Account: []string{accountAdmin},
	}

	iam, response, err := s.ControlPlane.CreateIAM(s.ctx, testUserID, policy)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.Require().NotNil(iam)

	testCases := []struct {
		testName      string
		policy        pds.ModelsAccessPolicy
		userID        string
		responseCode  int
		expectedError bool
	}{
		{
			testName:      "update IAM -- empty policy",
			policy:        pds.ModelsAccessPolicy{},
			userID:        testUserID,
			responseCode:  http.StatusUnprocessableEntity,
			expectedError: true,
		},
		{
			testName: "update IAM -- invalid tenant",
			policy: pds.ModelsAccessPolicy{
				Tenant: []pds.ModelsBinding{
					{
						RoleName:    &tenantAdmin,
						ResourceIds: []string{"invalid-tenant"},
					},
				},
			},
			userID:        testUserID,
			responseCode:  http.StatusBadRequest,
			expectedError: true,
		},
		{
			testName: "update IAM -- invalid tenant-role",
			policy: pds.ModelsAccessPolicy{
				Tenant: []pds.ModelsBinding{
					{
						RoleName:    &accountAdmin,
						ResourceIds: []string{tenantID},
					},
				},
			},
			userID:        testUserID,
			responseCode:  http.StatusUnprocessableEntity,
			expectedError: true,
		},
		{
			testName: "update IAM -- update should be successful",
			policy: pds.ModelsAccessPolicy{
				Tenant: []pds.ModelsBinding{
					{
						RoleName:    &tenantAdmin,
						ResourceIds: []string{tenantID},
					},
				},
			},
			userID:        testUserID,
			responseCode:  http.StatusOK,
			expectedError: false,
		},
		{
			testName: "update IAM -- global role should be successful",
			policy: pds.ModelsAccessPolicy{
				Global: []string{"pds-base"},
			},
			userID:        testUserID,
			responseCode:  http.StatusOK,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.testName, func(t *testing.T) {
			_, response, err := s.ControlPlane.UpdateIAM(s.ctx, tc.userID, tc.policy)
			s.checkError(err, tc.expectedError)
			s.Require().Equal(tc.responseCode, response.StatusCode)
		})
	}
	// cleanup.
	s.ControlPlane.MustDeleteIAM(s.ctx, s.T(), testUserID)
}

func (s *IAMTestSuite) checkError(err error, expectedError bool) {
	if expectedError {
		s.Require().Error(err)
	} else {
		s.Require().NoError(err)
	}
}

func (s *IAMTestSuite) testIAM_Update_UpdateByRemoving(tenantID, projectID string) {
	policy := pds.ModelsAccessPolicy{
		Account: []string{accountAdmin},
	}
	iam := s.ControlPlane.MustCreateIAM(s.ctx, s.T(), testUserID, policy)
	s.Require().NotNil(iam)
	s.Require().Equal(*iam.ActorId, testUserID)

	iam, response, err := s.ControlPlane.GetIAM(s.ctx, s.T(), testUserID)
	s.Require().NotNil(iam)
	s.Require().Equal(*iam.ActorId, testUserID)
	s.Require().NoError(err)
	s.Require().NotNil(response)
	s.Require().Equal(response.StatusCode, http.StatusOK)

	// Updating tenant role.
	policy.Account = iam.AccessPolicy.Account
	policy.Tenant = []pds.ModelsBinding{
		{
			RoleName:    &tenantAdmin,
			ResourceIds: []string{tenantID},
		},
	}
	iam = s.ControlPlane.MustUpdateIAM(s.ctx, s.T(), testUserID, policy)
	s.Require().NotNil(iam)
	s.Require().Equal(iam.AccessPolicy.Global, policy.Global)
	s.Require().Equal(iam.AccessPolicy.Account, policy.Account)
	s.Require().Equal(iam.AccessPolicy.Tenant, policy.Tenant)

	policy.Project = []pds.ModelsBinding{
		{
			RoleName:    &projectRole,
			ResourceIds: []string{projectID},
		},
	}
	iam = s.ControlPlane.MustUpdateIAM(s.ctx, s.T(), testUserID, policy)
	s.Require().NotNil(iam)
	s.Require().Equal(iam.AccessPolicy.Global, policy.Global)
	s.Require().Equal(iam.AccessPolicy.Account, policy.Account)
	s.Require().Equal(iam.AccessPolicy.Tenant, policy.Tenant)
	s.Require().Equal(iam.AccessPolicy.Project, policy.Project)

	// Updating by removing tenant roles from policy.
	policy.Tenant = []pds.ModelsBinding{}
	iam, response, err = s.ControlPlane.UpdateIAM(s.ctx, testUserID, policy)
	s.Require().NoError(err)
	s.Require().NotNil(iam)
	s.Require().NotNil(response)
	s.Require().Equal(response.StatusCode, http.StatusOK)
	s.Require().Equal(iam.AccessPolicy.Tenant, policy.Tenant)
	s.Require().Equal(iam.AccessPolicy.Account, policy.Account)

	// Updating by removing account roles from policy.
	policy.Project = []pds.ModelsBinding{}
	iam, response, err = s.ControlPlane.UpdateIAM(s.ctx, testUserID, policy)
	s.Require().NoError(err)
	s.Require().NotNil(iam)
	s.Require().NotNil(response)
	s.Require().Equal(response.StatusCode, http.StatusOK)

	// Deleting iam entry.
	s.ControlPlane.MustDeleteIAM(s.ctx, s.T(), testUserID)

	// Verifying delete.
	iam, response, err = s.ControlPlane.GetIAM(s.ctx, s.T(), testUserID)
	s.Require().Nil(iam)
	s.Require().Error(err)
	s.Require().Equal(response.StatusCode, http.StatusNotFound)
}

func (s *IAMTestSuite) testIAM_Update_GetList(tenantID string) {
	policy := pds.ModelsAccessPolicy{
		Account: []string{accountAdmin},
	}
	iam := s.ControlPlane.MustCreateIAM(s.ctx, s.T(), testUserID, policy)
	s.Require().NotNil(iam)
	s.Require().Equal(*iam.ActorId, testUserID)

	iam, response, err := s.ControlPlane.GetIAM(s.ctx, s.T(), testUserID)
	s.Require().NotNil(iam)
	s.Require().Equal(*iam.ActorId, testUserID)
	s.Require().Len(iam.AccessPolicy.Account, 1)
	s.Require().NoError(err)
	s.Require().NotNil(response)
	s.Require().Equal(response.StatusCode, http.StatusOK)

	// Updating with tenant role.
	policy.Account = iam.AccessPolicy.Account
	policy.Tenant = []pds.ModelsBinding{
		{
			RoleName:    &tenantAdmin,
			ResourceIds: []string{tenantID},
		},
	}
	iam = s.ControlPlane.MustUpdateIAM(s.ctx, s.T(), testUserID, policy)
	s.Require().NotNil(iam)
	s.Require().Equal(iam.AccessPolicy.Global, policy.Global)
	s.Require().Equal(iam.AccessPolicy.Account, policy.Account)
	s.Require().Equal(iam.AccessPolicy.Tenant, policy.Tenant)

	iams, response, err := s.ControlPlane.ListIAM(s.ctx, s.T())
	s.Require().NotEmpty(iams)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)

	var found bool
	for _, iam := range iams {
		if *iam.ActorId == testUserID {
			found = true
		}
	}
	s.Require().True(found)

	// Deleting iam entry.
	s.ControlPlane.MustDeleteIAM(s.ctx, s.T(), testUserID)

	// Verifying delete.
	iam, response, err = s.ControlPlane.GetIAM(s.ctx, s.T(), testUserID)
	s.Require().Nil(iam)
	s.Require().Error(err)
	s.Require().Equal(response.StatusCode, http.StatusNotFound)
}

func (s *IAMTestSuite) testIAM_VerifyAuth(tenantID, projectID string) {
	// Creating IAM for account roles.
	policy := pds.ModelsAccessPolicy{
		Account: []string{accountReader},
	}

	iam := s.ControlPlane.MustCreateIAM(s.ctx, s.T(), testUserID, policy)
	s.Require().NotNil(iam)
	s.Require().Equal(*iam.ActorId, testUserID)

	// fetching client for test auth user.
	pdsClient, err := s.getTestAuthUserPDSClient()
	s.Require().NoError(err)

	whoAmIPDSClient := s.ControlPlane.PDS
	s.ControlPlane.PDS = pdsClient

	credName := framework.NewRandomName("backupCredPrefix")
	accountKey := "Acc-key"
	accountName := "Acc-name"
	credentials := pds.ControllersCredentials{
		Azure: &pds.ModelsAzureCredentials{
			AccountKey:  &accountKey,
			AccountName: &accountName,
		},
	}

	// checking auth with test auth user.
	_, response, err := s.ControlPlane.CreateBackupCredentials(s.ctx, credName, credentials)
	s.Require().Error(err)
	s.Require().Equal(http.StatusForbidden, response.StatusCode)

	s.ControlPlane.PDS = whoAmIPDSClient
	// Updated the access for test auth user.
	policy.Account = []string{accountAdmin}
	_, response, err = s.ControlPlane.UpdateIAM(s.ctx, testUserID, policy)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)

	s.ControlPlane.PDS = pdsClient

	// Checking the access for test auth user.
	createdBackupCreds, response, err := s.ControlPlane.CreateBackupCredentials(s.ctx, credName, credentials)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.T().Cleanup(func() { s.ControlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), createdBackupCreds.GetId()) })

	s.ControlPlane.PDS = whoAmIPDSClient

	// Deleting IAM entry
	s.ControlPlane.MustDeleteIAM(s.ctx, s.T(), testUserID)

	// Verifying delete
	iam, response, err = s.ControlPlane.GetIAM(s.ctx, s.T(), testUserID)
	s.Require().Nil(iam)
	s.Require().Error(err)
	s.Require().Equal(response.StatusCode, http.StatusNotFound)
}
