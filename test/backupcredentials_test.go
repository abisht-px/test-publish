package test

import (
	"net/http"

	apiv1 "github.com/portworx/pds-api-go-client/pds/v1alpha1"
)

const backupCredPrefix = "backup-creds"

func (s *PDSTestSuite) TestBackupCredentials_CreateAndFetchCredentialsForMultipleObjectStores_Succeeded() {
	s3Creds := s.config.backupTarget.credentials.S3
	gcpCreds := "{\"creds\": \"fake-creds\"}"
	accKey := "accountKey"
	accName := "accountName"
	credName := generateRandomName(backupCredPrefix)

	testCases := []struct {
		description            string
		objectStoreCredentials apiv1.ControllersCredentials
	}{
		{
			description: "Google cloud storage",
			objectStoreCredentials: apiv1.ControllersCredentials{
				Google: &apiv1.ModelsGoogleCredentials{
					JsonKey:   &gcpCreds,
					ProjectId: &s.controlPlane.TestPDSProjectID,
				},
			},
		},
		{
			description: "Azure cloud object store",
			objectStoreCredentials: apiv1.ControllersCredentials{
				Azure: &apiv1.ModelsAzureCredentials{
					AccountKey:  &accKey,
					AccountName: &accName,
				},
			},
		},
		{
			description: "S3 object store",
			objectStoreCredentials: apiv1.ControllersCredentials{
				S3: &apiv1.ModelsS3Credentials{
					AccessKey: &s3Creds.AccessKey,
					Endpoint:  &s3Creds.Endpoint,
					SecretKey: &s3Creds.SecretKey,
				},
			},
		},
		{
			description: "S3 compatible object store",
			objectStoreCredentials: apiv1.ControllersCredentials{
				S3Compatible: &apiv1.ModelsS3CompatibleCredentials{
					AccessKey: &s3Creds.AccessKey,
					Endpoint:  &s3Creds.Endpoint,
					SecretKey: &s3Creds.SecretKey,
				},
			},
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.description, func() {
			createdBackupCreds := s.controlPlane.MustCreateBackupCredentials(s.ctx, s.T(), credName, testCase.objectStoreCredentials)
			s.T().Cleanup(func() { s.controlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), createdBackupCreds.GetId()) })

			backupCreds := s.controlPlane.MustGetBackupCredentials(s.ctx, s.T(), createdBackupCreds.GetId())
			s.Require().Equal(backupCreds.GetName(), credName)

			backupCredList := s.controlPlane.MustListBackupCredentials(s.ctx, s.T())
			s.Require().True(nameExistsInCredentialList(backupCredList, credName))

			cloudConfig := s.controlPlane.MustGetBackupCredentialsNoSecrets(s.ctx, s.T(), createdBackupCreds.GetId())
			switch {
			case cloudConfig.Google != nil:
				s.Require().NotNil(cloudConfig.Google)
				s.Require().Equal(cloudConfig.Google.GetProjectId(), s.controlPlane.TestPDSProjectID)
			case cloudConfig.Azure != nil:
				s.Require().NotNil(cloudConfig.Azure)
				s.Require().Equal(cloudConfig.Azure.GetAccountName(), accName)
			case cloudConfig.S3 != nil:
				s.Require().NotNil(cloudConfig.S3)
				s.Require().Equal(cloudConfig.S3.GetAccessKey(), s3Creds.AccessKey)
			case cloudConfig.S3Compatible != nil:
				s.Require().NotNil(cloudConfig.S3Compatible)
				s.Require().Equal(cloudConfig.S3Compatible.GetAccessKey(), s3Creds.AccessKey)
			default:
				s.Fail("On of the configurations need to be provided!")
			}
		})
	}
}

func (s *PDSTestSuite) TestBackupCredentials_DuplicateCredentialsCreation_ResultsInConflictError() {
	// Given.
	credName := generateRandomName(backupCredPrefix)
	accountKey := "Acc-key"
	accountName := "Acc-name"
	credentials := apiv1.ControllersCredentials{
		Azure: &apiv1.ModelsAzureCredentials{
			AccountKey:  &accountKey,
			AccountName: &accountName,
		},
	}

	createdBackupCreds := s.controlPlane.MustCreateBackupCredentials(s.ctx, s.T(), credName, credentials)
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), createdBackupCreds.GetId()) })

	// When.
	_, httpResponse, err := s.controlPlane.CreateBackupCredentials(s.ctx, credName, credentials)

	// Then.
	s.Require().Equal(http.StatusConflict, httpResponse.StatusCode)
	s.Require().Error(err)
}

func (s *PDSTestSuite) TestBackupCredentials_UpdateCredsNonAssociatedWithTarget_Succeeded() {
	// Given.
	credName := generateRandomName(backupCredPrefix)
	updatedName := generateRandomName("updated-" + backupCredPrefix)
	updatedJsonKey := "{\"creds\": \"fake-creds2\"}"
	createdBackupCreds := s.controlPlane.MustCreateGoogleBackupCredentials(s.ctx, s.T(), credName)
	s.T().Cleanup(func() { s.controlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), createdBackupCreds.GetId()) })

	backupCreds := s.controlPlane.MustGetBackupCredentials(s.ctx, s.T(), createdBackupCreds.GetId())
	s.Require().Equal(backupCreds.GetName(), credName)

	// When.
	s.controlPlane.MustUpdateGoogleBackupCredentials(s.ctx, s.T(), createdBackupCreds.GetId(), updatedName, updatedJsonKey)

	// Then.
	backupCreds = s.controlPlane.MustGetBackupCredentials(s.ctx, s.T(), createdBackupCreds.GetId())
	s.Require().Equal(backupCreds.GetName(), updatedName)
	backupCredList := s.controlPlane.MustListBackupCredentials(s.ctx, s.T())
	s.Require().True(nameExistsInCredentialList(backupCredList, updatedName))
	s.Require().False(nameExistsInCredentialList(backupCredList, credName))
}

func (s *PDSTestSuite) TestBackupCredentials_UpdateCredsAssociatedWithTarget_Failed() {
	s.T().Skip("DS-4819: The CP API is actually returning 200 now, but it shouldn't. Skipping until it's fixed.")
	// Given.
	credName := generateRandomName(backupCredPrefix)
	backupTargetConfig := s.config.backupTarget
	s3Creds := backupTargetConfig.credentials.S3
	updatedAccessKey := "BRANDNEWACCESSKEY"
	updatedCredentials := apiv1.ControllersCredentials{
		S3: &apiv1.ModelsS3Credentials{
			Endpoint:  &s3Creds.Endpoint,
			AccessKey: &updatedAccessKey,
			SecretKey: &s3Creds.SecretKey,
		},
	}

	backupCredentials := s.controlPlane.MustCreateS3BackupCredentials(s.ctx, s.T(), s3Creds, credName)
	backupTarget := s.controlPlane.MustCreateS3BackupTarget(s.ctx, s.T(), backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.T().Cleanup(func() {
		s.controlPlane.MustDeleteBackupTarget(s.ctx, s.T(), backupTarget.GetId())
		s.controlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), backupCredentials.GetId())
	})
	s.controlPlane.MustEnsureBackupTargetCreatedInTC(s.ctx, s.T(), backupTarget.GetId())

	// When.
	_, httpResponse, err := s.controlPlane.UpdateBackupCredentials(s.ctx, backupCredentials.GetId(), "new-name", updatedCredentials)

	// Then.
	s.Require().Equal(http.StatusConflict, httpResponse.StatusCode)
	s.Require().Error(err)
}

func (s *PDSTestSuite) TestBackupCredentials_DeleteCredsNonAssociatedWithTarget_Succeeded() {
	// Given.
	credName := generateRandomName(backupCredPrefix)
	createdBackupCreds := s.controlPlane.MustCreateGoogleBackupCredentials(s.ctx, s.T(), credName)

	// When.
	s.controlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), createdBackupCreds.GetId())

	// Then.
	_, httpResponse, _ := s.controlPlane.GetBackupCredentials(s.ctx, createdBackupCreds.GetId())
	s.Require().Equal(http.StatusNotFound, httpResponse.StatusCode)
}

func (s *PDSTestSuite) TestBackupCredentials_DeleteCredsAssociatedWithTarget_Failed() {
	// Given.
	credName := generateRandomName(backupCredPrefix)
	s3Creds := s.config.backupTarget.credentials.S3
	createdBackupCreds := s.controlPlane.MustCreateS3BackupCredentials(s.ctx, s.T(), s3Creds, credName)
	createdBackupTarget := s.controlPlane.MustCreateS3BackupTarget(s.ctx, s.T(), createdBackupCreds.GetId(), s.config.backupTarget.bucket, s.config.backupTarget.region)
	s.T().Cleanup(func() {
		s.controlPlane.MustDeleteBackupTarget(s.ctx, s.T(), createdBackupTarget.GetId())
		s.controlPlane.MustDeleteBackupCredentials(s.ctx, s.T(), createdBackupCreds.GetId())
	})

	// When.
	httpResponse, err := s.controlPlane.DeleteBackupCredentials(s.ctx, createdBackupCreds.GetId())

	// Then.
	s.Require().Equal(http.StatusConflict, httpResponse.StatusCode)
	s.Require().Error(err)
}

func nameExistsInCredentialList(backupCreds []apiv1.ModelsBackupCredentials, credName string) bool {
	for _, cred := range backupCreds {
		if cred.GetName() == credName {
			return true
		}
	}
	return false
}
