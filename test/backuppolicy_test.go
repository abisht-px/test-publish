package test

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/portworx/pds-integration-test/internal/random"
)

func (s *PDSTestSuite) TestBackupPolicy_CRUD_Ok() {
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-%s", nameSuffix)
	schedule := "* * * * 1"
	var retention int32 = 1
	backupPolicy := s.mustCreateBackupPolicy(&name, &schedule, &retention)
	s.T().Cleanup(func() {
		_, _ = s.deleteBackupPolicy(backupPolicy.GetId())
	})

	storedBackupPolicy := s.mustListBackupPolicy(backupPolicy.GetId())
	s.Require().Equal(storedBackupPolicy.Name, backupPolicy.Name)
	s.Require().Equal(storedBackupPolicy.Schedules[0].RetentionCount, backupPolicy.Schedules[0].RetentionCount)
	s.Require().Equal(storedBackupPolicy.Schedules[0].Schedule, backupPolicy.Schedules[0].Schedule)
	s.Require().Equal(storedBackupPolicy.Schedules[0].Id, backupPolicy.Schedules[0].Id)

	storedBackupPolicy = s.mustGetBackupPolicy(backupPolicy.GetId())
	s.Require().Equal(storedBackupPolicy.Name, backupPolicy.Name)
	s.Require().Equal(storedBackupPolicy.Schedules[0].RetentionCount, backupPolicy.Schedules[0].RetentionCount)
	s.Require().Equal(storedBackupPolicy.Schedules[0].Schedule, backupPolicy.Schedules[0].Schedule)
	s.Require().Equal(storedBackupPolicy.Schedules[0].Id, backupPolicy.Schedules[0].Id)

	newName := fmt.Sprintf("integration-test-updated-%s", nameSuffix)
	newSchedule := "* * * * 2"
	var newRetention int32 = 2
	backupPolicy = s.mustUpdateBackupPolicy(backupPolicy.GetId(), &newName, &newSchedule, &newRetention)
	s.Require().Equal(*backupPolicy.Name, newName)
	s.Require().Equal(*backupPolicy.Schedules[0].Schedule, newSchedule)
	s.Require().Equal(*backupPolicy.Schedules[0].RetentionCount, newRetention)

	s.mustDeleteBackupPolicy(backupPolicy.GetId())
}

func (s *PDSTestSuite) TestBackupPolicy_CreateDuplicateName_Conflict() {
	// Given.
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-%s", nameSuffix)
	schedule := "* * * * *"
	var retention int32 = 1
	backupPolicy := s.mustCreateBackupPolicy(&name, &schedule, &retention)
	// When.
	newBackupPolicy, resp, err := s.createBackupPolicy(&name, &schedule, &retention)
	s.T().Cleanup(func() {
		s.mustDeleteBackupPolicy(backupPolicy.GetId())
		// Clean BackupPolicy in case this tests accidentally creates a valid object.
		_, _ = s.deleteBackupPolicy(newBackupPolicy.GetId())
	})
	// Then.
	s.Require().Error(err)
	s.Require().Equal(http.StatusConflict, resp.StatusCode)
	s.Require().Nil(newBackupPolicy)
}

func (s *PDSTestSuite) TestBackupPolicy_CreateInvalidSchedule_Unprocessable() {
	// Given.
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-%s", nameSuffix)
	schedule := "a s d f g"
	var retention int32 = 1
	// When.
	backupPolicy, resp, err := s.createBackupPolicy(&name, &schedule, &retention)
	s.T().Cleanup(func() {
		// Clean BackupPolicy in case this tests accidentally creates a valid object.
		_, _ = s.deleteBackupPolicy(backupPolicy.GetId())
	})
	// Then.
	s.Require().Error(err)
	s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
	s.Require().Nil(backupPolicy)
}

func (s *PDSTestSuite) TestBackupPolicy_UpdateInvalidSchedule_Unprocessable() {
	// Given.
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-%s", nameSuffix)
	validSchedule := "* * * * *"
	invalidSchedule := "a s d f g"
	var retention int32 = 1
	backupPolicy := s.mustCreateBackupPolicy(&name, &validSchedule, &retention)
	s.T().Cleanup(func() {
		s.mustDeleteBackupPolicy(backupPolicy.GetId())
	})
	// When.
	updatedBackupPolicy, resp, err := s.updateBackupPolicy(backupPolicy.GetId(), &name, &invalidSchedule, &retention)
	// Then.
	s.Require().Error(err)
	s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
	s.Require().Nil(updatedBackupPolicy)
}

func (s *PDSTestSuite) TestBackupPolicy_DeleteNonExistent_NotFound() {
	// Given.
	id := uuid.New()
	// When.
	resp, err := s.deleteBackupPolicy(id.String())
	// Then.
	s.Require().Error(err)
	s.Require().Equal(http.StatusNotFound, resp.StatusCode)
}
