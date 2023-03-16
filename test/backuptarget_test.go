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
	backupTarget, response, err := s.createS3BackupTarget(backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.T().Cleanup(func() { s.deleteBackupTargetIfExists(backupTarget.GetId()) })

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
	backupTarget, response, err := s.createS3BackupTarget(backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.T().Cleanup(func() { s.deleteBackupTargetIfExists(backupTarget.GetId()) })

	// Then.
	api.RequireNoError(s.T(), response, err)
	s.mustWaitForBackupTargetState(s.T(), backupTarget.GetId(), s.testPDSDeploymentTargetID, "failed_create")
	backupTargetState := s.mustGetBackupTargetState(s.T(), backupTarget.GetId(), s.testPDSDeploymentTargetID)
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
	backupTarget := s.mustCreateS3BackupTarget(s.T(), backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.T().Cleanup(func() { s.deleteBackupTargetIfExists(backupTarget.GetId()) })

	// Then.
	// Check backup target state.
	s.mustEnsureBackupTargetCreatedInTC(s.T(), backupTarget.GetId(), s.testPDSDeploymentTargetID)
	backupTargetState := s.mustGetBackupTargetState(s.T(), backupTarget.GetId(), s.testPDSDeploymentTargetID)
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
	s.mustDeleteBackupTarget(s.T(), backupTarget.GetId())
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
