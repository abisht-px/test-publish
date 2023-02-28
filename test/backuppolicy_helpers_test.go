package test

import (
	"net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/test/api"
)

func (s *PDSTestSuite) mustCreateBackupPolicy(name, schedule *string, retention *int32) *pds.ModelsBackupPolicy {
	backupPolicy, resp, err := s.createBackupPolicy(name, schedule, retention)
	api.RequireNoError(s.T(), resp, err)
	return backupPolicy
}

func (s *PDSTestSuite) createBackupPolicy(name, schedule *string, retention *int32) (*pds.ModelsBackupPolicy, *http.Response, error) {
	policyType := "full"

	requestBody := pds.ControllersCreateBackupPolicyRequest{
		Name: name,
		Schedules: []pds.ModelsBackupSchedule{
			{
				Schedule:       schedule,
				RetentionCount: retention,
				Type:           &policyType,
			},
		},
	}
	return s.apiClient.BackupPoliciesApi.ApiTenantsIdBackupPoliciesPost(s.ctx, s.testPDSTenantID).Body(requestBody).Execute()
}

func (s *PDSTestSuite) mustListBackupPolicy(backupPolicyID string) *pds.ModelsBackupPolicy {
	backupPolicy, resp, err := s.apiClient.BackupPoliciesApi.ApiTenantsIdBackupPoliciesGet(s.ctx, s.testPDSTenantID).Id2(backupPolicyID).Execute()
	api.RequireNoError(s.T(), resp, err)
	s.Require().NotEmpty(backupPolicy)
	return &backupPolicy.Data[0]
}

func (s *PDSTestSuite) mustGetBackupPolicy(backupPolicyID string) *pds.ModelsBackupPolicy {
	backupPolicy, resp, err := s.apiClient.BackupPoliciesApi.ApiBackupPoliciesIdGet(s.ctx, backupPolicyID).Execute()
	api.RequireNoError(s.T(), resp, err)
	s.Require().NotNil(backupPolicy)
	return backupPolicy
}

func (s *PDSTestSuite) mustUpdateBackupPolicy(backupPolicyID string, name, schedule *string, retention *int32) *pds.ModelsBackupPolicy {
	backupPolicy, resp, err := s.updateBackupPolicy(backupPolicyID, name, schedule, retention)
	api.RequireNoError(s.T(), resp, err)
	return backupPolicy
}

func (s *PDSTestSuite) updateBackupPolicy(backupPolicyID string, name, schedule *string, retention *int32) (*pds.ModelsBackupPolicy, *http.Response, error) {
	policyType := "full"
	requestBody := pds.ControllersUpdateBackupPolicyRequest{
		Name: name,
		Schedules: []pds.ModelsBackupSchedule{
			{
				Schedule:       schedule,
				RetentionCount: retention,
				Type:           &policyType,
			},
		},
	}
	return s.apiClient.BackupPoliciesApi.ApiBackupPoliciesIdPut(s.ctx, backupPolicyID).Body(requestBody).Execute()
}

func (s *PDSTestSuite) mustDeleteBackupPolicy(backupPolicyID string) {
	resp, err := s.deleteBackupPolicy(backupPolicyID)
	api.RequireNoError(s.T(), resp, err)
}

func (s *PDSTestSuite) deleteBackupPolicy(backupPolicyID string) (*http.Response, error) {
	return s.apiClient.BackupPoliciesApi.ApiBackupPoliciesIdDelete(s.ctx, backupPolicyID).Execute()
}
