package controlplane

import (
	"context"
	"net/http"
	"testing"

	apiv1 "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
)

type BackupCredentials struct {
	S3 S3Credentials
}

type S3Credentials struct {
	AccessKey string
	Endpoint  string
	SecretKey string
}

func (c *ControlPlane) MustCreateS3BackupCredentials(ctx context.Context, t *testing.T, s3Creds S3Credentials, credName string) *apiv1.ModelsBackupCredentials {
	credentials := apiv1.ControllersCredentials{
		S3: &apiv1.ModelsS3Credentials{
			Endpoint:  &s3Creds.Endpoint,
			AccessKey: &s3Creds.AccessKey,
			SecretKey: &s3Creds.SecretKey,
		},
	}

	return c.MustCreateBackupCredentials(ctx, t, credName, credentials)
}

func (s *ControlPlane) MustCreateGoogleBackupCredentials(ctx context.Context, t *testing.T, credName string) *apiv1.ModelsBackupCredentials {
	myCreds := "{\"creds\": \"fake-creds\"}"
	credentials := apiv1.ControllersCredentials{
		Google: &apiv1.ModelsGoogleCredentials{
			JsonKey:   &myCreds,
			ProjectId: &s.TestPDSProjectID,
		},
	}

	return s.MustCreateBackupCredentials(ctx, t, credName, credentials)
}

func (s *ControlPlane) MustCreateBackupCredentials(ctx context.Context, t *testing.T, credName string, credentials apiv1.ControllersCredentials) *apiv1.ModelsBackupCredentials {
	backupCreds, httpResp, err := s.CreateBackupCredentials(ctx, credName, credentials)
	api.RequireNoError(t, httpResp, err)

	return backupCreds
}

func (s *ControlPlane) CreateBackupCredentials(ctx context.Context, credName string, credentials apiv1.ControllersCredentials) (*apiv1.ModelsBackupCredentials, *http.Response, error) {
	return s.PDS.BackupCredentialsApi.ApiTenantsIdBackupCredentialsPost(ctx, s.TestPDSTenantID).
		Body(apiv1.ControllersCreateBackupCredentialsRequest{Credentials: &credentials, Name: &credName}).
		Execute()
}

func (s *ControlPlane) MustGetBackupCredentials(ctx context.Context, t *testing.T, credentialsId string) *apiv1.ModelsBackupCredentials {
	backupCreds, httpResp, err := s.GetBackupCredentials(ctx, credentialsId)
	api.RequireNoError(t, httpResp, err)

	return backupCreds
}

func (s *ControlPlane) GetBackupCredentials(ctx context.Context, credentialsId string) (*apiv1.ModelsBackupCredentials, *http.Response, error) {
	return s.PDS.BackupCredentialsApi.ApiBackupCredentialsIdGet(ctx, credentialsId).Execute()
}

func (s *ControlPlane) MustGetBackupCredentialsNoSecrets(ctx context.Context, t *testing.T, credentialsId string) *apiv1.ControllersPartialCredentials {
	cloudConfig, httpResp, err := s.PDS.BackupCredentialsApi.ApiBackupCredentialsIdCredentialsGet(ctx, credentialsId).Execute()
	api.RequireNoError(t, httpResp, err)

	return cloudConfig
}

func (s *ControlPlane) MustListBackupCredentials(ctx context.Context, t *testing.T) []apiv1.ModelsBackupCredentials {
	backupCredList, httpResp, err := s.PDS.BackupCredentialsApi.ApiTenantsIdBackupCredentialsGet(ctx, s.TestPDSTenantID).Execute()
	api.RequireNoError(t, httpResp, err)

	return backupCredList.GetData()
}

func (s *ControlPlane) MustUpdateGoogleBackupCredentials(ctx context.Context, t *testing.T, credentialsId string, name string, jsonKey string) *apiv1.ModelsBackupCredentials {
	backupCreds, httpResp, err := s.updateGoogleBackupCredentials(ctx, credentialsId, name, jsonKey)

	api.RequireNoError(t, httpResp, err)

	return backupCreds
}

func (s *ControlPlane) updateGoogleBackupCredentials(ctx context.Context, credentialsId string, name string, jsonKey string) (*apiv1.ModelsBackupCredentials, *http.Response, error) {
	credentials := apiv1.ControllersCredentials{
		Google: &apiv1.ModelsGoogleCredentials{
			JsonKey:   &jsonKey,
			ProjectId: &s.TestPDSProjectID,
		},
	}

	return s.UpdateBackupCredentials(ctx, credentialsId, name, credentials)
}

func (s *ControlPlane) UpdateBackupCredentials(ctx context.Context, credentialsId string, name string, credentials apiv1.ControllersCredentials) (*apiv1.ModelsBackupCredentials, *http.Response, error) {
	return s.PDS.BackupCredentialsApi.ApiBackupCredentialsIdPut(ctx, credentialsId).
		Body(apiv1.ControllersUpdateBackupCredentialsRequest{Credentials: &credentials, Name: &name}).
		Execute()
}

func (s *ControlPlane) MustDeleteBackupCredentials(ctx context.Context, t *testing.T, backupCredentialsID string) {
	resp, err := s.DeleteBackupCredentials(ctx, backupCredentialsID)
	api.RequireNoError(t, resp, err)
}

func (s *ControlPlane) DeleteBackupCredentialsIfExists(ctx context.Context, t *testing.T, backupCredentialsID string) {
	resp, err := s.DeleteBackupCredentials(ctx, backupCredentialsID)
	if resp.StatusCode == http.StatusNotFound {
		return
	}
	api.NoError(t, resp, err)
}

func (s *ControlPlane) DeleteBackupCredentials(ctx context.Context, backupCredentialsID string) (*http.Response, error) {
	return s.PDS.BackupCredentialsApi.ApiBackupCredentialsIdDelete(ctx, backupCredentialsID).Execute()
}
