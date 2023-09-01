package backup_test

import (
	"net/http"

	apiv1 "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/suites/framework"
)

const backupCredPrefix = "backup-creds"

func (s *BackupTestSuite) TestBackupCredentials_CreateAndFetchCredentialsForMultipleObjectStores_Succeeded() {
	s3Creds := backupTargetCfg.Credentials.S3
	gcpCreds := "{\"creds\": \"fake-creds\"}"
	accKey := "accountKey"
	accName := "accountName"
	credName := framework.NewRandomName(backupCredPrefix)

	testCases := []struct {
		description            string
		objectStoreCredentials apiv1.ControllersCredentials
	}{
		{
			description: "Google cloud storage",
			objectStoreCredentials: apiv1.ControllersCredentials{
				Google: &apiv1.ModelsGoogleCredentials{
					JsonKey:   &gcpCreds,
					ProjectId: &controlPlane.TestPDSProjectID,
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
			createdBackupCreds := controlPlane.MustCreateBackupCredentials(ctx, s.T(), credName, testCase.objectStoreCredentials)
			s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), createdBackupCreds.GetId()) })

			backupCreds := controlPlane.MustGetBackupCredentials(ctx, s.T(), createdBackupCreds.GetId())
			s.Require().Equal(backupCreds.GetName(), credName)

			backupCredList := controlPlane.MustListBackupCredentials(ctx, s.T())
			s.Require().True(nameExistsInCredentialList(backupCredList, credName))

			cloudConfig := controlPlane.MustGetBackupCredentialsNoSecrets(ctx, s.T(), createdBackupCreds.GetId())
			switch {
			case cloudConfig.Google != nil:
				s.Require().NotNil(cloudConfig.Google)
				s.Require().Equal(cloudConfig.Google.GetProjectId(), controlPlane.TestPDSProjectID)
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

func (s *BackupTestSuite) TestBackupCredentials_DuplicateCredentialsCreation_ResultsInConflictError() {
	// Given.
	credName := framework.NewRandomName(backupCredPrefix)
	accountKey := "Acc-key"
	accountName := "Acc-name"
	credentials := apiv1.ControllersCredentials{
		Azure: &apiv1.ModelsAzureCredentials{
			AccountKey:  &accountKey,
			AccountName: &accountName,
		},
	}

	createdBackupCreds := controlPlane.MustCreateBackupCredentials(ctx, s.T(), credName, credentials)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), createdBackupCreds.GetId()) })

	// When.
	_, httpResponse, err := controlPlane.CreateBackupCredentials(ctx, credName, credentials)

	// Then.
	s.Require().Equal(http.StatusConflict, httpResponse.StatusCode)
	s.Require().Error(err)
}

func (s *BackupTestSuite) TestBackupCredentials_UpdateCredsNonAssociatedWithTarget_Succeeded() {
	// Given.
	credName := framework.NewRandomName(backupCredPrefix)
	updatedName := framework.NewRandomName("updated-" + backupCredPrefix)
	updatedJsonKey := "{\"creds\": \"fake-creds2\"}"
	createdBackupCreds := controlPlane.MustCreateGoogleBackupCredentials(ctx, s.T(), credName)
	s.T().Cleanup(func() { controlPlane.MustDeleteBackupCredentials(ctx, s.T(), createdBackupCreds.GetId()) })

	backupCreds := controlPlane.MustGetBackupCredentials(ctx, s.T(), createdBackupCreds.GetId())
	s.Require().Equal(backupCreds.GetName(), credName)

	// When.
	controlPlane.MustUpdateGoogleBackupCredentials(ctx, s.T(), createdBackupCreds.GetId(), updatedName, updatedJsonKey)

	// Then.
	backupCreds = controlPlane.MustGetBackupCredentials(ctx, s.T(), createdBackupCreds.GetId())
	s.Require().Equal(backupCreds.GetName(), updatedName)
	backupCredList := controlPlane.MustListBackupCredentials(ctx, s.T())
	s.Require().True(nameExistsInCredentialList(backupCredList, updatedName))
	s.Require().False(nameExistsInCredentialList(backupCredList, credName))
}

func (s *BackupTestSuite) TestBackupCredentials_UpdateCredsAssociatedWithTarget_Failed() {
	s.T().Skip("DS-4819: The CP API is actually returning 200 now, but it shouldn't. Skipping until it's fixed.")
	// Given.
	credName := framework.NewRandomName(backupCredPrefix)
	s3Creds := backupTargetCfg.Credentials.S3
	updatedAccessKey := "BRANDNEWACCESSKEY"
	updatedCredentials := apiv1.ControllersCredentials{
		S3: &apiv1.ModelsS3Credentials{
			Endpoint:  &s3Creds.Endpoint,
			AccessKey: &updatedAccessKey,
			SecretKey: &s3Creds.SecretKey,
		},
	}

	backupCredentials := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, credName)
	backupTarget := controlPlane.MustCreateS3BackupTarget(ctx, s.T(), backupCredentials.GetId(), backupTargetCfg.Bucket, backupTargetCfg.Region)
	s.T().Cleanup(func() {
		controlPlane.MustDeleteBackupTarget(ctx, s.T(), backupTarget.GetId())
		controlPlane.MustDeleteBackupCredentials(ctx, s.T(), backupCredentials.GetId())
	})
	controlPlane.MustEnsureBackupTargetCreatedInTC(ctx, s.T(), backupTarget.GetId())

	// When.
	_, httpResponse, err := controlPlane.UpdateBackupCredentials(ctx, backupCredentials.GetId(), "new-name", updatedCredentials)

	// Then.
	s.Require().Equal(http.StatusConflict, httpResponse.StatusCode)
	s.Require().Error(err)
}

func (s *BackupTestSuite) TestBackupCredentials_DeleteCredsNonAssociatedWithTarget_Succeeded() {
	// Given.
	credName := framework.NewRandomName(backupCredPrefix)
	createdBackupCreds := controlPlane.MustCreateGoogleBackupCredentials(ctx, s.T(), credName)

	// When.
	controlPlane.MustDeleteBackupCredentials(ctx, s.T(), createdBackupCreds.GetId())

	// Then.
	_, httpResponse, _ := controlPlane.GetBackupCredentials(ctx, createdBackupCreds.GetId())
	s.Require().Equal(http.StatusNotFound, httpResponse.StatusCode)
}

func (s *BackupTestSuite) TestBackupCredentials_DeleteCredsAssociatedWithTarget_Failed() {
	// Given.
	credName := framework.NewRandomName(backupCredPrefix)
	s3Creds := backupTargetCfg.Credentials.S3
	createdBackupCreds := controlPlane.MustCreateS3BackupCredentials(ctx, s.T(), s3Creds, credName)
	createdBackupTarget := controlPlane.MustCreateS3BackupTarget(
		ctx, s.T(),
		createdBackupCreds.GetId(),
		backupTargetCfg.Bucket,
		backupTargetCfg.Region,
	)
	s.T().Cleanup(func() {
		controlPlane.MustDeleteBackupTarget(ctx, s.T(), createdBackupTarget.GetId())
		controlPlane.MustDeleteBackupCredentials(ctx, s.T(), createdBackupCreds.GetId())
	})

	// When.
	httpResponse, err := controlPlane.DeleteBackupCredentials(ctx, createdBackupCreds.GetId())

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
