package test

import (
	"net/http"

	apiv1 "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
)

// Helper functions performing CRUD operations on Control Plane API.

func (s *PDSTestSuite) mustCreateS3BackupCredentials(s3Creds s3Credentials, credName string) *apiv1.ModelsBackupCredentials {
	credentials := apiv1.ControllersCredentials{
		S3: &apiv1.ModelsS3Credentials{
			Endpoint:  &s3Creds.endpoint,
			AccessKey: &s3Creds.accessKey,
			SecretKey: &s3Creds.secretKey,
		},
	}

	return s.mustCreateBackupCredentials(credName, credentials)
}

func (s *PDSTestSuite) mustCreateGoogleBackupCredentials(credName string) *apiv1.ModelsBackupCredentials {
	myCreds := "{\"creds\": \"fake-creds\"}"
	credentials := apiv1.ControllersCredentials{
		Google: &apiv1.ModelsGoogleCredentials{
			JsonKey:   &myCreds,
			ProjectId: &s.testPDSProjectID,
		},
	}

	return s.mustCreateBackupCredentials(credName, credentials)
}

func (s *PDSTestSuite) mustCreateBackupCredentials(credName string, credentials apiv1.ControllersCredentials) *apiv1.ModelsBackupCredentials {
	backupCreds, httpResp, err := s.createBackupCredentials(credName, credentials)
	api.RequireNoError(s.T(), httpResp, err)

	return backupCreds
}

func (s *PDSTestSuite) createBackupCredentials(credName string, credentials apiv1.ControllersCredentials) (*apiv1.ModelsBackupCredentials, *http.Response, error) {
	return s.apiClient.BackupCredentialsApi.ApiTenantsIdBackupCredentialsPost(s.ctx, s.testPDSTenantID).
		Body(apiv1.ControllersCreateBackupCredentialsRequest{Credentials: &credentials, Name: &credName}).
		Execute()
}

func (s *PDSTestSuite) mustGetBackupCredentials(credentialsId string) *apiv1.ModelsBackupCredentials {
	backupCreds, httpResp, err := s.getBackupCredentials(credentialsId)
	api.RequireNoError(s.T(), httpResp, err)

	return backupCreds
}

func (s *PDSTestSuite) getBackupCredentials(credentialsId string) (*apiv1.ModelsBackupCredentials, *http.Response, error) {
	return s.apiClient.BackupCredentialsApi.ApiBackupCredentialsIdGet(s.ctx, credentialsId).Execute()
}

func (s *PDSTestSuite) mustGetBackupCredentialsNoSecrets(credentialsId string) *apiv1.ControllersPartialCredentials {
	cloudConfig, httpResp, err := s.apiClient.BackupCredentialsApi.ApiBackupCredentialsIdCredentialsGet(s.ctx, credentialsId).Execute()
	api.RequireNoError(s.T(), httpResp, err)

	return cloudConfig
}

func (s *PDSTestSuite) mustListBackupCredentials() []apiv1.ModelsBackupCredentials {
	backupCredList, httpResp, err := s.apiClient.BackupCredentialsApi.ApiTenantsIdBackupCredentialsGet(s.ctx, s.testPDSTenantID).Execute()
	api.RequireNoError(s.T(), httpResp, err)

	return backupCredList.GetData()
}

func (s *PDSTestSuite) mustUpdateGoogleBackupCredentials(credentialsId string, name string, jsonKey string) *apiv1.ModelsBackupCredentials {
	backupCreds, httpResp, err := s.updateGoogleBackupCredentials(credentialsId, name, jsonKey)

	api.RequireNoError(s.T(), httpResp, err)

	return backupCreds
}

func (s *PDSTestSuite) updateGoogleBackupCredentials(credentialsId string, name string, jsonKey string) (*apiv1.ModelsBackupCredentials, *http.Response, error) {
	credentials := apiv1.ControllersCredentials{
		Google: &apiv1.ModelsGoogleCredentials{
			JsonKey:   &jsonKey,
			ProjectId: &s.testPDSProjectID,
		},
	}

	return s.updateBackupCredentials(credentialsId, name, credentials)
}

func (s *PDSTestSuite) updateBackupCredentials(credentialsId string, name string, credentials apiv1.ControllersCredentials) (*apiv1.ModelsBackupCredentials, *http.Response, error) {
	return s.apiClient.BackupCredentialsApi.ApiBackupCredentialsIdPut(s.ctx, credentialsId).
		Body(apiv1.ControllersUpdateBackupCredentialsRequest{Credentials: &credentials, Name: &name}).
		Execute()
}

func (s *PDSTestSuite) mustDeleteBackupCredentials(backupCredentialsID string) {
	resp, err := s.deleteBackupCredentials(backupCredentialsID)
	api.RequireNoError(s.T(), resp, err)
}

func (s *PDSTestSuite) deleteBackupCredentialsIfExists(backupCredentialsID string) {
	resp, err := s.deleteBackupCredentials(backupCredentialsID)
	if resp.StatusCode == http.StatusNotFound {
		return
	}
	api.NoError(s.T(), resp, err)
}

func (s *PDSTestSuite) deleteBackupCredentials(backupCredentialsID string) (*http.Response, error) {
	return s.apiClient.BackupCredentialsApi.ApiBackupCredentialsIdDelete(s.ctx, backupCredentialsID).Execute()
}
