package test

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/portworx"
)

func (s *PDSTestSuite) TestBackupTarget_EmptyBackupTarget_Fail() {
	// Given.
	backupTargetConfig := backupTargetConfig{
		credentials: s.config.backupTarget.credentials,
	}
	backupCredentialsConfig := backupTargetConfig.credentials.s3
	backupCredentials := s.mustCreateS3BackupCredentials(s.T(), backupCredentialsConfig, generateRandomName(backupCredPrefix))
	s.T().Cleanup(func() { s.deleteBackupCredentialsIfExists(backupCredentials.GetId()) })

	// When.
	backupTarget, response, err := s.controlPlane.CreateS3BackupTarget(s.ctx, backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.T().Cleanup(func() { s.controlPlane.DeleteBackupTargetIfExists(s.ctx, s.T(), backupTarget.GetId()) })

	// Then.
	s.Require().Equal(http.StatusUnprocessableEntity, response.StatusCode)
	s.Require().Error(err)
	s.Require().Nil(backupTarget)
}

func (s *PDSTestSuite) TestBackupTarget_InvalidNonemptyBackupTarget_Fail() {
	// Given.
	backupTargetConfig := backupTargetConfig{
		bucket: "xxx",
		region: "xxx",
		credentials: backupCredentials{
			s3: s3Credentials{
				accessKey: "xxx",
				endpoint:  "xxx",
				secretKey: "xxx",
			},
		},
	}

	backupCredentialsConfig := backupTargetConfig.credentials.s3
	backupCredentials := s.mustCreateS3BackupCredentials(s.T(), backupCredentialsConfig, generateRandomName(backupCredPrefix))
	s.T().Cleanup(func() { s.deleteBackupCredentialsIfExists(backupCredentials.GetId()) })

	// When.
	backupTarget, response, err := s.controlPlane.CreateS3BackupTarget(s.ctx, backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.T().Cleanup(func() { s.controlPlane.DeleteBackupTargetIfExists(s.ctx, s.T(), backupTarget.GetId()) })

	// Then.
	api.RequireNoError(s.T(), response, err)
	s.controlPlane.MustWaitForBackupTargetState(s.ctx, s.T(), backupTarget.GetId(), "failed_create")
	backupTargetState := s.controlPlane.MustGetBackupTargetState(s.ctx, s.T(), backupTarget.GetId())
	s.Require().NotEmpty(backupTargetState.GetErrorDetails())
	s.Require().NotEmpty(backupTargetState.GetErrorMessage())
	s.Require().Equal("failed_to_create_px_credentials", backupTargetState.GetErrorCode())
	s.Require().Empty(backupTargetState.GetPxCredentialsName())
	s.Require().Equal(uuid.Nil.String(), backupTargetState.GetPxCredentialsId())
}

func (s *PDSTestSuite) TestBackupTarget_CreateAndDeleteInTC_Succeed() {
	// Given.
	backupTargetConfig := s.config.backupTarget
	backupCredentialsConfig := backupTargetConfig.credentials.s3

	backupCredentials := s.mustCreateS3BackupCredentials(s.T(), backupCredentialsConfig, generateRandomName(backupCredPrefix))
	s.T().Cleanup(func() { s.deleteBackupCredentialsIfExists(backupCredentials.GetId()) })

	// When.
	backupTarget := s.controlPlane.MustCreateS3BackupTarget(s.ctx, s.T(), backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.T().Cleanup(func() { s.controlPlane.DeleteBackupTargetIfExists(s.ctx, s.T(), backupTarget.GetId()) })

	// Then.
	// Check backup target state.
	s.controlPlane.MustEnsureBackupTargetCreatedInTC(s.ctx, s.T(), backupTarget.GetId())
	backupTargetState := s.controlPlane.MustGetBackupTargetState(s.ctx, s.T(), backupTarget.GetId())
	s.Require().Empty(backupTargetState.GetErrorCode())
	s.Require().Empty(backupTargetState.GetErrorDetails())
	s.Require().Empty(backupTargetState.GetErrorMessage())
	s.Require().NotEmpty(backupTargetState.GetPxCredentialsName())
	s.Require().NotEmpty(backupTargetState.GetPxCredentialsId())

	// Check PX Cloud credentials in the target cluster.
	pxCloudCredentials, err := s.targetCluster.ListPXCloudCredentials(s.ctx)
	s.Require().NoError(err)
	foundPXCloudCredential := s.findCloudCredentialByName(pxCloudCredentials, backupTargetState.GetPxCredentialsName())
	s.Require().NotNil(foundPXCloudCredential)
	s.Require().Equal(backupTargetState.GetPxCredentialsId(), foundPXCloudCredential.ID)
	s.Require().Equal(backupTargetConfig.bucket, foundPXCloudCredential.Bucket)
	s.Require().Equal(backupTargetConfig.region, foundPXCloudCredential.AwsCredential.Region)
	s.Require().Equal(backupCredentialsConfig.accessKey, foundPXCloudCredential.AwsCredential.AccessKey)
	s.Require().Equal(backupCredentialsConfig.endpoint, foundPXCloudCredential.AwsCredential.Endpoint)

	// Test deletion of the backup target.
	s.controlPlane.MustDeleteBackupTarget(s.ctx, s.T(), backupTarget.GetId())
	pxCloudCredentials, err = s.targetCluster.ListPXCloudCredentials(s.ctx)
	s.Require().NoError(err)
	foundPXCloudCredential = s.findCloudCredentialByName(pxCloudCredentials, backupTargetState.GetPxCredentialsName())
	s.Require().Nil(foundPXCloudCredential)
}

func (s *PDSTestSuite) findCloudCredentialByName(pxCloudCredentials []portworx.PXCloudCredential, name string) *portworx.PXCloudCredential {
	for _, pxCloudCredential := range pxCloudCredentials {
		if pxCloudCredential.Name == name {
			return &pxCloudCredential
		}
	}
	return nil
}
