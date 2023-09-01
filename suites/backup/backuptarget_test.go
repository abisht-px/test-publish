package backup_test

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/portworx"
	"github.com/portworx/pds-integration-test/suites/framework"
)

func (s *BackupTestSuite) TestBackupTarget_EmptyBackupTarget_Fail() {
	// Given.
	backupCredentialsConfig := backupTargetCfg.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), backupCredentialsConfig, framework.NewRandomName(backupCredPrefix))
	s.T().Cleanup(func() { controlPlane.DeleteBackupCredentialsIfExists(ctx, s.T(), backupCredentials.GetId()) })

	// When.
	backupTarget, response, err := controlPlane.CreateS3BackupTarget(ctx, backupCredentials.GetId(), "", backupTargetCfg.Region)
	s.T().Cleanup(func() { controlPlane.DeleteBackupTargetIfExists(ctx, s.T(), backupTarget.GetId()) })

	// Then.
	s.Require().Equal(http.StatusUnprocessableEntity, response.StatusCode)
	s.Require().Error(err)
	s.Require().Nil(backupTarget)
}

func (s *BackupTestSuite) TestBackupTarget_InvalidNonemptyBackupTarget_Fail() {
	// Given.
	backupTargetConfig := framework.BackupTargetConfig{
		Bucket: "xxx",
		Region: "xxx",
		Credentials: controlplane.BackupCredentials{
			S3: controlplane.S3Credentials{
				AccessKey: "xxx",
				Endpoint:  "xxx",
				SecretKey: "xxx",
			},
		},
	}

	backupCredentialsConfig := backupTargetConfig.Credentials.S3
	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), backupCredentialsConfig, framework.NewRandomName(backupCredPrefix))
	s.T().Cleanup(func() { controlPlane.DeleteBackupCredentialsIfExists(ctx, s.T(), backupCredentials.GetId()) })

	// When.
	backupTarget, response, err := controlPlane.CreateS3BackupTarget(ctx, backupCredentials.GetId(), backupTargetConfig.Bucket, backupTargetConfig.Region)
	s.T().Cleanup(func() { controlPlane.DeleteBackupTargetIfExists(ctx, s.T(), backupTarget.GetId()) })

	// Then.
	api.RequireNoError(s.T(), response, err)
	controlPlane.MustWaitForBackupTargetState(ctx, s.T(), backupTarget.GetId(), "failed_create")
	backupTargetState := controlPlane.MustGetBackupTargetState(ctx, s.T(), backupTarget.GetId())
	s.Require().NotEmpty(backupTargetState.GetErrorDetails())
	s.Require().NotEmpty(backupTargetState.GetErrorMessage())
	s.Require().Equal("failed_to_create_px_credentials", backupTargetState.GetErrorCode())
	s.Require().Empty(backupTargetState.GetPxCredentialsName())
	s.Require().Equal(uuid.Nil.String(), backupTargetState.GetPxCredentialsId())
}

func (s *BackupTestSuite) TestBackupTarget_CreateAndDeleteInTC_Succeed() {
	// Given.
	backupTargetConfig := backupTargetCfg
	backupCredentialsConfig := backupTargetConfig.Credentials.S3

	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), backupCredentialsConfig, framework.NewRandomName(backupCredPrefix))
	s.T().Cleanup(func() { controlPlane.DeleteBackupCredentialsIfExists(ctx, s.T(), backupCredentials.GetId()) })

	// When.
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetCfg.Bucket, backupTargetCfg.Region)
	s.T().Cleanup(func() { controlPlane.DeleteBackupTargetIfExists(ctx, s.T(), backupTarget.GetId()) })

	// Then.
	// Check backup target state.
	backupTargetState := pds.ModelsBackupTargetState{}

	s.Run(fmt.Sprintf("Check Backup Target State for ID:%s", backupTarget.GetId()), func() {
		controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())
		backupTargetState = controlPlane.MustGetBackupTargetState(ctx, s.T(), backupTarget.GetId())
		s.Require().Empty(backupTargetState.GetErrorCode())
		s.Require().Empty(backupTargetState.GetErrorDetails())
		s.Require().Empty(backupTargetState.GetErrorMessage())
		s.Require().NotEmpty(backupTargetState.GetPxCredentialsName())
		s.Require().NotEmpty(backupTargetState.GetPxCredentialsId())
	})

	// Check PX Cloud credentials in the target cluster.
	s.Run("Check PX Cloud credentials in the target cluster", func() {
		pxCloudCredentials, err := targetCluster.ListPXCloudCredentials(ctx)
		s.Require().NoError(err)
		foundPXCloudCredential := s.findCloudCredentialByName(pxCloudCredentials, backupTargetState.GetPxCredentialsName())
		s.Require().NotNil(foundPXCloudCredential)
		s.Require().Equal(backupTargetState.GetPxCredentialsId(), foundPXCloudCredential.ID)
		s.Require().Equal(backupTargetCfg.Bucket, foundPXCloudCredential.Bucket)
		s.Require().Equal(backupTargetCfg.Region, foundPXCloudCredential.AwsCredential.Region)
		s.Require().Equal(backupCredentialsConfig.AccessKey, foundPXCloudCredential.AwsCredential.AccessKey)
		s.Require().Equal(backupCredentialsConfig.Endpoint, foundPXCloudCredential.AwsCredential.Endpoint)
	})

	s.Run("Test deletion of the backup target", func() {
		controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId())
		pxCloudCredentials, err := targetCluster.ListPXCloudCredentials(ctx)
		s.Require().NoError(err)
		foundPXCloudCredential := s.findCloudCredentialByName(pxCloudCredentials, backupTargetState.GetPxCredentialsName())
		s.Require().Nil(foundPXCloudCredential)
	})
}

func (s *BackupTestSuite) findCloudCredentialByName(pxCloudCredentials []portworx.PXCloudCredential, name string) *portworx.PXCloudCredential {
	for _, pxCloudCredential := range pxCloudCredentials {
		if pxCloudCredential.Name == name {
			return &pxCloudCredential
		}
	}
	return nil
}
