package iam_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/portworx/pds-integration-test/internal/api"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/pointer"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/random"
)

func (s *IAMTestSuite) Test_ServiceIdentity_Create() {
	//
	serviceIdentity, response, err := s.ControlPlane.CreateServiceIdentity(s.ctx, s.ControlPlane.TestPDSAccountID, "service-identity"+random.AlphaNumericString(5), true)
	s.Require().NoError(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.T().Cleanup(func() {
		_, err := s.ControlPlane.DeleteServiceIdentity(s.ctx, serviceIdentity.GetId())
		require.NoError(s.T(), err)
	})
	enable := false
	tests := []struct {
		TestName        string
		ServiceIdentity *pds.ModelsServiceIdentityWithToken
		Request         *pds.RequestsServiceIdentityRequest
		AccountID       string

		ResponseCode int
	}{
		{
			TestName:  "Create_Invalid",
			AccountID: s.ControlPlane.TestPDSAccountID,
			Request: &pds.RequestsServiceIdentityRequest{
				Name:        "",
				Description: nil,
				Enabled:     enable,
			},
			ResponseCode: http.StatusBadRequest,
		},
		{
			TestName:  "Create_Duplicate",
			AccountID: s.ControlPlane.TestPDSAccountID,
			Request: &pds.RequestsServiceIdentityRequest{
				Name:        *serviceIdentity.Name,
				Description: serviceIdentity.Description,
				Enabled:     *serviceIdentity.Enabled,
			},
			ResponseCode: http.StatusConflict,
		},
		{
			TestName:  "Create_Account_NotFound",
			AccountID: uuid.New().String(),
			Request: &pds.RequestsServiceIdentityRequest{
				Name:        "test-name",
				Description: pointer.String("description"),
				Enabled:     enable,
			},
			ResponseCode: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		s.T().Run(test.TestName, func(t *testing.T) {
			_, response, _ := s.ControlPlane.CreateServiceIdentity(s.ctx, test.AccountID, test.Request.Name, test.Request.Enabled)
			s.Require().Equal(test.ResponseCode, response.StatusCode)
		})
	}
}

func (s *IAMTestSuite) Test_ServiceIdentity_Get() {
	//
	serviceIdentity, response, err := s.ControlPlane.CreateServiceIdentity(s.ctx, s.ControlPlane.TestPDSAccountID, "service-identity"+random.AlphaNumericString(5), true)
	s.Require().NoError(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.T().Cleanup(func() {
		_, err := s.ControlPlane.DeleteServiceIdentity(s.ctx, serviceIdentity.GetId())
		require.NoError(s.T(), err)
	})

	tests := []struct {
		TestName        string
		ServiceIdentity pds.ModelsServiceIdentityWithToken

		ResponseCode int
	}{
		{
			TestName:        "FindByID_Ok",
			ServiceIdentity: *serviceIdentity,
			ResponseCode:    http.StatusOK,
		},
		{
			TestName:        "FindByID_NotFound",
			ServiceIdentity: pds.ModelsServiceIdentityWithToken{Id: pointer.String(uuid.NewString())},
			ResponseCode:    http.StatusNotFound,
		},
	}

	for _, test := range tests {
		s.T().Run(test.TestName, func(t *testing.T) {
			result, resp, err := s.ControlPlane.GetServiceIdentity(s.ctx, s.T(), test.ServiceIdentity.GetId())

			s.Require().Equal(test.ResponseCode, resp.StatusCode)

			if test.ResponseCode != http.StatusOK {
				s.Require().Error(err)
				return
			}
			s.Require().Equal(test.ServiceIdentity.Id, result.Id)
			s.Require().Equal(test.ServiceIdentity.Name, result.Name)
			s.Require().Equal(test.ServiceIdentity.Description, result.Description)
			s.Require().Equal(test.ServiceIdentity.Enabled, result.Enabled)
			s.Require().Equal(test.ServiceIdentity.AccountId, result.AccountId)
			s.Require().Equal(test.ServiceIdentity.ClientId, result.ClientId)
			s.Require().Equal(test.ServiceIdentity.SecretGenerationCount, result.SecretGenerationCount)

		})

	}
}

func (s *IAMTestSuite) Test_ServiceIdentity_Update() {
	serviceIdentity, response, err := s.ControlPlane.CreateServiceIdentity(s.ctx, s.ControlPlane.TestPDSAccountID, "service-identity"+random.AlphaNumericString(5), true)
	s.Require().NoError(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.T().Cleanup(func() {
		_, err := s.ControlPlane.DeleteServiceIdentity(s.ctx, serviceIdentity.GetId())
		require.NoError(s.T(), err)
	})
	enable := false
	tests := []struct {
		TestName        string
		ServiceIdentity *pds.ModelsServiceIdentityWithToken
		Request         *pds.RequestsServiceIdentityRequest

		ResponseCode int
	}{
		{
			TestName:        "Update_Ok",
			ServiceIdentity: serviceIdentity,
			Request: &pds.RequestsServiceIdentityRequest{
				Name:        "updated-name" + random.AlphaNumericString(5),
				Description: pointer.String("updated description"),
				Enabled:     enable,
			},
			ResponseCode: http.StatusOK,
		},
		{
			TestName:        "Update_Invalid",
			ServiceIdentity: serviceIdentity,
			Request: &pds.RequestsServiceIdentityRequest{
				Name:        "",
				Description: nil,
				Enabled:     enable,
			},
			ResponseCode: http.StatusBadRequest,
		},
		{
			TestName:        "Update_NotFound",
			ServiceIdentity: &pds.ModelsServiceIdentityWithToken{Id: pointer.String(uuid.NewString())},
			Request: &pds.RequestsServiceIdentityRequest{
				Name:        "updated-name",
				Description: pointer.String("updated description"),
				Enabled:     enable,
			},
			ResponseCode: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		s.T().Run(test.TestName, func(t *testing.T) {
			response, _ := s.ControlPlane.UpdateServiceIdentity(s.ctx, test.ServiceIdentity.GetId(), test.Request)
			s.Require().Equal(test.ResponseCode, response.StatusCode)
		})
	}
}

func (s *IAMTestSuite) Test_ServiceIdentity_Delete() {
	serviceIdentity, response, err := s.ControlPlane.CreateServiceIdentity(s.ctx, s.ControlPlane.TestPDSAccountID, "service-identity"+random.AlphaNumericString(5), true)
	s.Require().NoError(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusOK, response.StatusCode)

	tests := []struct {
		TestName        string
		ServiceIdentity *pds.ModelsServiceIdentityWithToken

		ResponseCode int
	}{
		{
			TestName:        "DeleteByID_Ok",
			ServiceIdentity: serviceIdentity,
			ResponseCode:    http.StatusNoContent,
		},
		{
			TestName:        "DeleteByID_NotFound",
			ServiceIdentity: &pds.ModelsServiceIdentityWithToken{Id: pointer.String(uuid.New().String())},
			ResponseCode:    http.StatusNotFound,
		},
	}

	for _, test := range tests {
		s.T().Run(test.TestName, func(t *testing.T) {
			resp, _ := s.ControlPlane.DeleteServiceIdentity(s.ctx, test.ServiceIdentity.GetId())
			s.Require().Equal(test.ResponseCode, resp.StatusCode)
		})
	}
}

func (s *IAMTestSuite) Test_ServiceIdentity_Regenerate() {
	serviceIdentity, response, err := s.ControlPlane.CreateServiceIdentity(s.ctx, s.ControlPlane.TestPDSAccountID, "service-identity"+random.AlphaNumericString(5), true)
	s.Require().NoError(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.T().Cleanup(func() {
		_, err := s.ControlPlane.DeleteServiceIdentity(s.ctx, serviceIdentity.GetId())
		require.NoError(s.T(), err)
	})

	tests := []struct {
		TestName        string
		ServiceIdentity *pds.ModelsServiceIdentityWithToken

		ResponseCode int
	}{
		{
			TestName:        "Regenerate_Ok",
			ServiceIdentity: serviceIdentity,
			ResponseCode:    http.StatusOK,
		},
		{
			TestName:        "Regenerate_NotFound",
			ServiceIdentity: &pds.ModelsServiceIdentityWithToken{Id: pointer.String(uuid.New().String())},
			ResponseCode:    http.StatusNotFound,
		},
	}

	for _, test := range tests {
		s.T().Run(test.TestName, func(t *testing.T) {
			result, resp, _ := s.ControlPlane.RegenerateServiceIdentity(s.ctx, test.ServiceIdentity.GetId())

			s.Require().Equal(test.ResponseCode, resp.StatusCode)
			if test.ResponseCode != http.StatusOK {
				return
			}
			s.Require().Equal(test.ServiceIdentity.Id, result.Id)
			s.Require().Equal(test.ServiceIdentity.Name, result.Name)
			s.Require().Equal(test.ServiceIdentity.Description, result.Description)
			s.Require().Equal(test.ServiceIdentity.Enabled, result.Enabled)
			s.Require().Equal(test.ServiceIdentity.AccountId, result.AccountId)
			s.Require().NotEqual(test.ServiceIdentity.ClientId, result.ClientId)
			s.Require().NotEqual(test.ServiceIdentity.ClientToken, result.ClientToken)
			s.Require().Equal(test.ServiceIdentity.GetSecretGenerationCount()+1, *result.SecretGenerationCount)

		})
	}
}

func (s *IAMTestSuite) Test_ServiceIdentity_GenerateToken() {
	serviceIdentity, response, err := s.ControlPlane.CreateServiceIdentity(s.ctx, s.ControlPlane.TestPDSAccountID, "service-identity"+random.AlphaNumericString(5), true)
	s.Require().NoError(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.T().Cleanup(func() {
		_, err := s.ControlPlane.DeleteServiceIdentity(s.ctx, serviceIdentity.GetId())
		require.NoError(s.T(), err)
	})

	tests := []struct {
		TestName        string
		ServiceIdentity *pds.ModelsServiceIdentityWithToken
		Payload         *pds.ControllersGenerateTokenRequest
		ResponseCode    int
	}{
		{
			TestName:        "Regenerate_Ok",
			ServiceIdentity: serviceIdentity,
			Payload: &pds.ControllersGenerateTokenRequest{ClientId: serviceIdentity.ClientId,
				ClientToken: serviceIdentity.ClientToken},
			ResponseCode: http.StatusOK,
		},
		{
			TestName:        "GenerateToken_InvalidClientID",
			ServiceIdentity: serviceIdentity,
			Payload: &pds.ControllersGenerateTokenRequest{ClientId: pointer.String(ClientID(s.T())),
				ClientToken: serviceIdentity.ClientToken},
			ResponseCode: http.StatusUnprocessableEntity,
		},
		{
			TestName:        "GenerateToken_InvalidClientSecret",
			ServiceIdentity: serviceIdentity,
			Payload: &pds.ControllersGenerateTokenRequest{ClientId: serviceIdentity.ClientId,
				ClientToken: pointer.String(ClientSecret(s.T()))},
			ResponseCode: http.StatusUnprocessableEntity,
		},
	}

	for _, test := range tests {
		s.T().Run(test.TestName, func(t *testing.T) {
			token, resp, _ := s.ControlPlane.GenerateTokenServiceIdentity(s.ctx, test.Payload)
			s.Require().Equal(test.ResponseCode, resp.StatusCode)
			if test.ResponseCode != http.StatusOK {
				return
			}
			serviceClient, err := api.NewPDSClient(s.ctx, s.ControlPlane.PDS.URL, api.LoginCredentials{BearerToken: *token.Token})
			s.Require().NoError(err)
			// Get the info about the auth user.
			authServiceResponse, response, err := serviceClient.WhoAmIApi.ApiWhoamiGet(context.Background()).Execute()
			api.RequireNoError(s.T(), response, err)
			result, userOk := authServiceResponse.GetServiceIdentityOk()
			s.Require().True(userOk)
			s.Require().NotNil(result.GetId())
			// Try the token: get accounts.
			_, response, _ = serviceClient.AccountsApi.ApiAccountsIdGet(s.ctx, s.ControlPlane.TestPDSAccountID).Execute()
			s.Require().Equal(http.StatusNotFound, response.StatusCode)

		})
	}
}

func (s *IAMTestSuite) Test_ServiceIdentity_List() {
	var serviceIdentities []*pds.ModelsServiceIdentityWithToken
	for i := 0; i < 5; i++ {
		result, _, _ := s.ControlPlane.CreateServiceIdentity(s.ctx, s.ControlPlane.TestPDSAccountID, "service-identity-test-list"+random.AlphaNumericString(5), true)
		serviceIdentities = append(
			serviceIdentities,
			result)
	}
	s.T().Cleanup(func() {
		count := len(serviceIdentities)
		for i := 0; i < count; i++ {
			_, err := s.ControlPlane.DeleteServiceIdentity(s.ctx, serviceIdentities[i].GetId())
			require.NoError(s.T(), err)
		}
	})
	tests := []struct {
		// Test Meta.
		TestName string

		// Input.
		APi pds.ApiApiAccountsIdServiceIdentityGetRequest
		// Output Validator
		Validator       func(a *pds.ModelsServiceIdentityWithToken, b *pds.ModelsServiceIdentity) bool
		ServiceIdentity *pds.ModelsServiceIdentityWithToken
		NumRecords      int
	}{
		{
			TestName: "ListByAccountID_FilterName",
			APi:      s.ControlPlane.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(s.ctx, s.ControlPlane.TestPDSAccountID).Name(*serviceIdentities[0].Name),
			Validator: func(a *pds.ModelsServiceIdentityWithToken, b *pds.ModelsServiceIdentity) bool {
				return strings.Compare(*a.Name, *b.Name) == 0
			},
			ServiceIdentity: serviceIdentities[0],
			NumRecords:      1,
		},
		{
			TestName: "ListByAccountID_SortByCreatedAtWithLimit",
			Validator: func(a *pds.ModelsServiceIdentityWithToken, b *pds.ModelsServiceIdentity) bool {
				return *a.CreatedAt > *(b.CreatedAt)
			},
			APi:             s.ControlPlane.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(s.ctx, s.ControlPlane.TestPDSAccountID).SortBy("created_at").Limit("3"),
			NumRecords:      3,
			ServiceIdentity: serviceIdentities[3],
		},
		{
			TestName: "ListByAccountID_FilterByID",
			Validator: func(a *pds.ModelsServiceIdentityWithToken, b *pds.ModelsServiceIdentity) bool {
				return strings.Compare(*a.Id, *b.Id) == 0
			},
			APi:             s.ControlPlane.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(s.ctx, s.ControlPlane.TestPDSAccountID).Id2(*serviceIdentities[4].Id),
			ServiceIdentity: serviceIdentities[4],
			NumRecords:      1,
		},
		{
			TestName: "ListByAccountID_FilterByCreatedBy",
			Validator: func(a *pds.ModelsServiceIdentityWithToken, b *pds.ModelsServiceIdentity) bool {
				return strings.Compare(*a.CreatedBy, *b.CreatedBy) == 0
			},
			APi:             s.ControlPlane.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(s.ctx, s.ControlPlane.TestPDSAccountID).CreatedBy(*serviceIdentities[3].CreatedBy),
			ServiceIdentity: serviceIdentities[0],
			NumRecords:      5,
		},
		{
			TestName: "ListByAccountID_FilterByClientID",
			Validator: func(a *pds.ModelsServiceIdentityWithToken, b *pds.ModelsServiceIdentity) bool {
				return strings.Compare(*a.ClientId, *b.ClientId) == 0

			},
			APi:             s.ControlPlane.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(s.ctx, s.ControlPlane.TestPDSAccountID).ClientId(*serviceIdentities[2].ClientId),
			ServiceIdentity: serviceIdentities[2],
			NumRecords:      1,
		},
		{
			TestName: "ListByAccountID_FilterByEnabled_True",
			Validator: func(a *pds.ModelsServiceIdentityWithToken, b *pds.ModelsServiceIdentity) bool {
				return *a.Enabled == *b.Enabled
			},
			APi:             s.ControlPlane.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(s.ctx, s.ControlPlane.TestPDSAccountID).Enabled(true).Limit("5"),
			ServiceIdentity: serviceIdentities[2],
			NumRecords:      5,
		},
		{
			TestName: "ListByAccountID_FilterByEnabled_False",
			Validator: func(a *pds.ModelsServiceIdentityWithToken, b *pds.ModelsServiceIdentity) bool {
				return *a.Enabled == *b.Enabled
			},
			APi:             s.ControlPlane.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(s.ctx, s.ControlPlane.TestPDSAccountID).Enabled(false),
			ServiceIdentity: serviceIdentities[0],
			NumRecords:      0,
		},
		{
			TestName: "ListByAccountID_FilterByEnabled_False",
			Validator: func(a *pds.ModelsServiceIdentityWithToken, b *pds.ModelsServiceIdentity) bool {
				return *a.Enabled == *b.Enabled
			},
			APi:             s.ControlPlane.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(s.ctx, s.ControlPlane.TestPDSAccountID).Enabled(false),
			ServiceIdentity: serviceIdentities[0],
			NumRecords:      0,
		},
	}

	for _, test := range tests {
		s.T().Run(test.TestName, func(t *testing.T) {
			serviceIdentities, _, err := test.APi.Execute()
			s.Require().NoError(err)
			s.Require().Equal(test.NumRecords, len(serviceIdentities.Data))
			for i := 0; i < test.NumRecords; i++ {
				s.Require().True(test.Validator(test.ServiceIdentity, &serviceIdentities.Data[i]))
			}
		})
	}
}

func (s *IAMTestSuite) Test_ServiceIdentity_With_IAM() {
	serviceIdentity, response, err := s.ControlPlane.CreateServiceIdentity(s.ctx, s.ControlPlane.TestPDSAccountID, "service-identity-list"+random.AlphaNumericString(5), true)
	s.Require().NoError(err)
	s.Require().NotNil(response)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.T().Cleanup(func() {
		_, err := s.ControlPlane.DeleteServiceIdentity(s.ctx, serviceIdentity.GetId())
		require.NoError(s.T(), err)
	})
	payload := &pds.ControllersGenerateTokenRequest{ClientId: serviceIdentity.ClientId,
		ClientToken: serviceIdentity.ClientToken}
	token, resp, _ := s.ControlPlane.GenerateTokenServiceIdentity(s.ctx, payload)
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	serviceClient, err := api.NewPDSClient(context.Background(), s.ControlPlane.PDS.URL, api.LoginCredentials{BearerToken: *token.Token})
	s.Require().NoError(err)
	// Try the token: get accounts.
	_, response, _ = serviceClient.AccountsApi.ApiAccountsIdGet(context.Background(), s.ControlPlane.TestPDSAccountID).Execute()
	s.Require().Equal(http.StatusNotFound, response.StatusCode)

	// Creating IAM for account roles.
	iam := s.ControlPlane.MustCreateIAM(s.ctx, s.T(), *serviceIdentity.Id, pds.ModelsAccessPolicy{
		Account: []string{"account-admin"},
	})
	s.Require().NotNil(iam)
	s.Require().Equal(iam.ActorId, serviceIdentity.Id)
	// Try the token: get accounts
	_, response, _ = serviceClient.AccountsApi.ApiAccountsIdGet(context.Background(), s.ControlPlane.TestPDSAccountID).Execute()
	s.Require().Equal(http.StatusOK, response.StatusCode)

	// Update role for serviceIdentity and verify
	policy := new(pds.ModelsAccessPolicy)
	projectRole := "project-admin"
	policy.Project = []pds.ModelsBinding{
		{
			RoleName:    &projectRole,
			ResourceIds: []string{s.ControlPlane.TestPDSProjectID},
		},
	}

	updatedIAM := s.ControlPlane.MustUpdateIAM(s.ctx, s.T(), serviceIdentity.GetId(), *policy)
	s.Require().NotNil(updatedIAM)
	s.Require().Equal(updatedIAM.ActorId, serviceIdentity.Id)
	// Get Tenant
	_, resp, _ = serviceClient.TenantsApi.ApiTenantsIdGet(context.Background(), s.ControlPlane.TestPDSAccountID).Execute()
	s.Require().Equal(http.StatusNotFound, resp.StatusCode)
	// Get Account
	_, response, _ = serviceClient.AccountsApi.ApiAccountsIdGet(context.Background(), s.ControlPlane.TestPDSAccountID).Execute()
	s.Require().Equal(http.StatusForbidden, response.StatusCode)

	// Get Project
	_, response, _ = serviceClient.ProjectsApi.ApiProjectsIdGet(context.Background(), s.ControlPlane.TestPDSTenantID).Execute()
	s.Require().Equal(http.StatusNotFound, response.StatusCode)

}

func (s *IAMTestSuite) Test_ServiceIdentity_With_Pagination() {
	var serviceIdentities []*pds.ModelsServiceIdentityWithToken
	for i := 0; i < 5; i++ {
		result, _, _ := s.ControlPlane.CreateServiceIdentity(s.ctx, s.ControlPlane.TestPDSAccountID, "service-identity-test-list"+random.AlphaNumericString(5), true)
		serviceIdentities = append(
			serviceIdentities,
			result)
	}
	s.T().Cleanup(func() {
		count := len(serviceIdentities)
		for i := 0; i < count; i++ {
			_, err := s.ControlPlane.DeleteServiceIdentity(s.ctx, serviceIdentities[i].GetId())
			require.NoError(s.T(), err)
		}
	})

	paginatedResult, resp, err := s.ControlPlane.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(s.ctx, s.ControlPlane.TestPDSAccountID).SortBy("created_by").Limit("3").Execute()
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal(3, len(paginatedResult.GetData()))
	s.Require().NotNil(paginatedResult.Pagination)
	s.Require().NotEmpty(paginatedResult.Pagination.Continuation)
	result, _, _ := s.ControlPlane.PDS.ServiceIdentityApi.ApiAccountsIdServiceIdentityGet(s.ctx, s.ControlPlane.TestPDSAccountID).SortBy("created_by").Limit("2").Execute()
	s.Require().Equal(2, len(result.GetData()))
	s.Require().NotNil(result.Pagination)
}

func ClientID(t *testing.T) string {
	return MustGetRandomString(t, "", 24)
}

func ClientSecret(t *testing.T) string {
	return MustGetRandomString(t, "", 36)
}

func MustGetRandomString(t *testing.T, prefix string, suffixLength int) string {
	t.Helper()
	bytes := make([]byte, suffixLength)
	_, err := rand.Read(bytes)
	require.NoError(t, err, "Generating random string suffix")
	hash := hex.EncodeToString(bytes)
	return fmt.Sprintf("%s-%s", prefix, hash[:suffixLength])
}
