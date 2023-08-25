package backup_test

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/portworx/pds-integration-test/internal/random"
)

func (s *BackupTestSuite) TestBackupPolicy_CRUD_Ok() {
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-%s", nameSuffix)
	schedule := "* * * * 1"
	var retention int32 = 1
	backupPolicy := controlPlane.MustCreateBackupPolicy(ctx, s.T(), &name, &schedule, &retention)
	s.T().Cleanup(func() {
		_, _ = controlPlane.DeleteBackupPolicy(ctx, backupPolicy.GetId())
	})

	storedBackupPolicy := controlPlane.MustListBackupPolicy(ctx, s.T(), backupPolicy.GetId())
	s.Require().Equal(storedBackupPolicy.Name, backupPolicy.Name)
	s.Require().Equal(storedBackupPolicy.Schedules[0].RetentionCount, backupPolicy.Schedules[0].RetentionCount)
	s.Require().Equal(storedBackupPolicy.Schedules[0].Schedule, backupPolicy.Schedules[0].Schedule)
	s.Require().Equal(storedBackupPolicy.Schedules[0].Id, backupPolicy.Schedules[0].Id)

	storedBackupPolicy = controlPlane.MustGetBackupPolicy(ctx, s.T(), backupPolicy.GetId())
	s.Require().Equal(storedBackupPolicy.Name, backupPolicy.Name)
	s.Require().Equal(storedBackupPolicy.Schedules[0].RetentionCount, backupPolicy.Schedules[0].RetentionCount)
	s.Require().Equal(storedBackupPolicy.Schedules[0].Schedule, backupPolicy.Schedules[0].Schedule)
	s.Require().Equal(storedBackupPolicy.Schedules[0].Id, backupPolicy.Schedules[0].Id)

	newName := fmt.Sprintf("integration-test-updated-%s", nameSuffix)
	newSchedule := "* * * * 2"
	var newRetention int32 = 2
	backupPolicy = controlPlane.MustUpdateBackupPolicy(ctx, s.T(), backupPolicy.GetId(), &newName, &newSchedule, &newRetention)
	s.Require().Equal(*backupPolicy.Name, newName)
	s.Require().Equal(*backupPolicy.Schedules[0].Schedule, newSchedule)
	s.Require().Equal(*backupPolicy.Schedules[0].RetentionCount, newRetention)

	controlPlane.MustDeleteBackupPolicy(ctx, s.T(), backupPolicy.GetId())
}

func (s *BackupTestSuite) TestBackupPolicy_CreateDuplicateName_Conflict() {
	// Given.
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-%s", nameSuffix)
	schedule := "* * * * *"
	var retention int32 = 1
	backupPolicy := controlPlane.MustCreateBackupPolicy(ctx, s.T(), &name, &schedule, &retention)
	// When.
	newBackupPolicy, resp, err := controlPlane.CreateBackupPolicy(ctx, &name, &schedule, &retention)
	s.T().Cleanup(func() {
		controlPlane.MustDeleteBackupPolicy(ctx, s.T(), backupPolicy.GetId())
		// Clean BackupPolicy in case this tests accidentally creates a valid object.
		_, _ = controlPlane.DeleteBackupPolicy(ctx, newBackupPolicy.GetId())
	})
	// Then.
	s.Require().Error(err)
	s.Require().Equal(http.StatusConflict, resp.StatusCode)
	s.Require().Nil(newBackupPolicy)
}

func (s *BackupTestSuite) TestBackupPolicy_CreateInvalidSchedule_Unprocessable() {
	// Given.
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-%s", nameSuffix)
	schedule := "a s d f g"
	var retention int32 = 1
	// When.
	backupPolicy, resp, err := controlPlane.CreateBackupPolicy(ctx, &name, &schedule, &retention)
	s.T().Cleanup(func() {
		// Clean BackupPolicy in case this tests accidentally creates a valid object.
		_, _ = controlPlane.DeleteBackupPolicy(ctx, backupPolicy.GetId())
	})
	// Then.
	s.Require().Error(err)
	s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
	s.Require().Nil(backupPolicy)
}

func (s *BackupTestSuite) TestBackupPolicy_UpdateInvalidSchedule_Unprocessable() {
	// Given.
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-%s", nameSuffix)
	validSchedule := "* * * * *"
	invalidSchedule := "a s d f g"
	var retention int32 = 1
	backupPolicy := controlPlane.MustCreateBackupPolicy(ctx, s.T(), &name, &validSchedule, &retention)
	s.T().Cleanup(func() {
		controlPlane.MustDeleteBackupPolicy(ctx, s.T(), backupPolicy.GetId())
	})
	// When.
	updatedBackupPolicy, resp, err := controlPlane.UpdateBackupPolicy(ctx, backupPolicy.GetId(), &name, &invalidSchedule, &retention)
	// Then.
	s.Require().Error(err)
	s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
	s.Require().Nil(updatedBackupPolicy)
}

func (s *BackupTestSuite) TestBackupPolicy_DeleteNonExistent_NotFound() {
	// Given.
	id := uuid.New()
	// When.
	resp, err := controlPlane.DeleteBackupPolicy(ctx, id.String())
	// Then.
	s.Require().Error(err)
	s.Require().Equal(http.StatusNotFound, resp.StatusCode)
}
