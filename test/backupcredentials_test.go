package test

import (
	"net/http"

	apiv1 "github.com/portworx/pds-api-go-client/pds/v1alpha1"
)

const backupCredPrefix = "backup-creds"

func (s *PDSTestSuite) TestBackupCredentials_CreateAndFetchCredentialsForMultipleObjectStores_Succeeded() {
	s3Creds := s.config.backupTarget.credentials.s3
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
					ProjectId: &s.testPDSProjectID,
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
					AccessKey: &s3Creds.accessKey,
					Endpoint:  &s3Creds.endpoint,
					SecretKey: &s3Creds.secretKey,
				},
			},
		},
		{
			description: "S3 compatible object store",
			objectStoreCredentials: apiv1.ControllersCredentials{
				S3Compatible: &apiv1.ModelsS3CompatibleCredentials{
					AccessKey: &s3Creds.accessKey,
					Endpoint:  &s3Creds.endpoint,
					SecretKey: &s3Creds.secretKey,
				},
			},
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.description, func() {
			createdBackupCreds := s.mustCreateBackupCredentials(s.T(), credName, testCase.objectStoreCredentials)
			s.T().Cleanup(func() { s.mustDeleteBackupCredentials(s.T(), createdBackupCreds.GetId()) })

			backupCreds := s.mustGetBackupCredentials(createdBackupCreds.GetId())
			s.Require().Equal(backupCreds.GetName(), credName)

			backupCredList := s.mustListBackupCredentials()
			s.Require().True(nameExistsInCredentialList(backupCredList, credName))

			cloudConfig := s.mustGetBackupCredentialsNoSecrets(createdBackupCreds.GetId())
			switch {
			case cloudConfig.Google != nil:
				s.Require().NotNil(cloudConfig.Google)
				s.Require().Equal(cloudConfig.Google.GetProjectId(), s.testPDSProjectID)
			case cloudConfig.Azure != nil:
				s.Require().NotNil(cloudConfig.Azure)
				s.Require().Equal(cloudConfig.Azure.GetAccountName(), accName)
			case cloudConfig.S3 != nil:
				s.Require().NotNil(cloudConfig.S3)
				s.Require().Equal(cloudConfig.S3.GetAccessKey(), s3Creds.accessKey)
			case cloudConfig.S3Compatible != nil:
				s.Require().NotNil(cloudConfig.S3Compatible)
				s.Require().Equal(cloudConfig.S3Compatible.GetAccessKey(), s3Creds.accessKey)
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

	createdBackupCreds := s.mustCreateBackupCredentials(s.T(), credName, credentials)
	s.T().Cleanup(func() { s.mustDeleteBackupCredentials(s.T(), createdBackupCreds.GetId()) })

	// When.
	_, httpResponse, err := s.createBackupCredentials(credName, credentials)

	// Then.
	s.Require().Equal(http.StatusConflict, httpResponse.StatusCode)
	s.Require().Error(err)
}

func (s *PDSTestSuite) TestBackupCredentials_UpdateCredsNonAssociatedWithTarget_Succeeded() {
	// Given.
	credName := generateRandomName(backupCredPrefix)
	updatedName := generateRandomName("updated-" + backupCredPrefix)
	updatedJsonKey := "{\"creds\": \"fake-creds2\"}"
	createdBackupCreds := s.mustCreateGoogleBackupCredentials(s.T(), credName)
	s.T().Cleanup(func() { s.mustDeleteBackupCredentials(s.T(), createdBackupCreds.GetId()) })

	backupCreds := s.mustGetBackupCredentials(createdBackupCreds.GetId())
	s.Require().Equal(backupCreds.GetName(), credName)

	// When.
	s.mustUpdateGoogleBackupCredentials(createdBackupCreds.GetId(), updatedName, updatedJsonKey)

	// Then.
	backupCreds = s.mustGetBackupCredentials(createdBackupCreds.GetId())
	s.Require().Equal(backupCreds.GetName(), updatedName)
	backupCredList := s.mustListBackupCredentials()
	s.Require().True(nameExistsInCredentialList(backupCredList, updatedName))
	s.Require().False(nameExistsInCredentialList(backupCredList, credName))
}

func (s *PDSTestSuite) TestBackupCredentials_UpdateCredsAssociatedWithTarget_Failed() {
	s.T().Skip("DS-4819: The CP API is actually returning 200 now, but it shouldn't. Skipping until it's fixed.")
	// Given.
	credName := generateRandomName(backupCredPrefix)
	backupTargetConfig := s.config.backupTarget
	s3Creds := backupTargetConfig.credentials.s3
	updatedAccessKey := "BRANDNEWACCESSKEY"
	updatedCredentials := apiv1.ControllersCredentials{
		S3: &apiv1.ModelsS3Credentials{
			Endpoint:  &s3Creds.endpoint,
			AccessKey: &updatedAccessKey,
			SecretKey: &s3Creds.secretKey,
		},
	}

	backupCredentials := s.mustCreateS3BackupCredentials(s.T(), s3Creds, credName)
	backupTarget := s.mustCreateS3BackupTarget(s.T(), backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
	s.T().Cleanup(func() {
		s.mustDeleteBackupTarget(s.T(), backupTarget.GetId())
		s.mustDeleteBackupCredentials(s.T(), backupCredentials.GetId())
	})
	s.mustEnsureBackupTargetCreatedInTC(s.T(), backupTarget.GetId(), s.testPDSDeploymentTargetID)

	// When.
	_, httpResponse, err := s.updateBackupCredentials(backupCredentials.GetId(), "new-name", updatedCredentials)

	// Then.
	s.Require().Equal(http.StatusConflict, httpResponse.StatusCode)
	s.Require().Error(err)
}

func (s *PDSTestSuite) TestBackupCredentials_DeleteCredsNonAssociatedWithTarget_Succeeded() {
	// Given.
	credName := generateRandomName(backupCredPrefix)
	createdBackupCreds := s.mustCreateGoogleBackupCredentials(s.T(), credName)

	// When.
	s.mustDeleteBackupCredentials(s.T(), createdBackupCreds.GetId())

	// Then.
	_, httpResponse, _ := s.getBackupCredentials(createdBackupCreds.GetId())
	s.Require().Equal(http.StatusNotFound, httpResponse.StatusCode)
}

func (s *PDSTestSuite) TestBackupCredentials_DeleteCredsAssociatedWithTarget_Failed() {
	// Given.
	credName := generateRandomName(backupCredPrefix)
	s3Creds := s.config.backupTarget.credentials.s3
	createdBackupCreds := s.mustCreateS3BackupCredentials(s.T(), s3Creds, credName)
	createdBackupTarget := s.mustCreateS3BackupTarget(s.T(), createdBackupCreds.GetId(), s.config.backupTarget.bucket, s.config.backupTarget.region)
	s.T().Cleanup(func() {
		s.mustDeleteBackupTarget(s.T(), createdBackupTarget.GetId())
		s.mustDeleteBackupCredentials(s.T(), createdBackupCreds.GetId())
	})

	// When.
	httpResponse, err := s.deleteBackupCredentials(createdBackupCreds.GetId())

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
