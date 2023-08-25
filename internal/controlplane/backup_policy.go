package controlplane

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/require"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) MustCreateBackupPolicy(ctx context.Context, t tests.T, name, schedule *string, retention *int32) *pds.ModelsBackupPolicy {
	backupPolicy, resp, err := c.CreateBackupPolicy(ctx, name, schedule, retention)
	api.RequireNoError(t, resp, err)
	return backupPolicy
}

func (c *ControlPlane) CreateBackupPolicy(ctx context.Context, name, schedule *string, retention *int32) (*pds.ModelsBackupPolicy, *http.Response, error) {
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
	return c.PDS.BackupPoliciesApi.ApiTenantsIdBackupPoliciesPost(ctx, c.TestPDSTenantID).Body(requestBody).Execute()
}

func (c *ControlPlane) MustListBackupPolicy(ctx context.Context, t tests.T, backupPolicyID string) *pds.ModelsBackupPolicy {
	backupPolicy, resp, err := c.PDS.BackupPoliciesApi.ApiTenantsIdBackupPoliciesGet(ctx, c.TestPDSTenantID).Id2(backupPolicyID).Execute()
	api.RequireNoError(t, resp, err)
	require.NotEmpty(t, backupPolicy)
	return &backupPolicy.Data[0]
}

func (c *ControlPlane) MustGetBackupPolicy(ctx context.Context, t tests.T, backupPolicyID string) *pds.ModelsBackupPolicy {
	backupPolicy, resp, err := c.PDS.BackupPoliciesApi.ApiBackupPoliciesIdGet(ctx, backupPolicyID).Execute()
	api.RequireNoError(t, resp, err)
	require.NotNil(t, backupPolicy)
	return backupPolicy
}

func (c *ControlPlane) MustUpdateBackupPolicy(ctx context.Context, t tests.T, backupPolicyID string, name, schedule *string, retention *int32) *pds.ModelsBackupPolicy {
	backupPolicy, resp, err := c.UpdateBackupPolicy(ctx, backupPolicyID, name, schedule, retention)
	api.RequireNoError(t, resp, err)
	return backupPolicy
}

func (c *ControlPlane) UpdateBackupPolicy(ctx context.Context, backupPolicyID string, name, schedule *string, retention *int32) (*pds.ModelsBackupPolicy, *http.Response, error) {
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
	return c.PDS.BackupPoliciesApi.ApiBackupPoliciesIdPut(ctx, backupPolicyID).Body(requestBody).Execute()
}

func (c *ControlPlane) MustDeleteBackupPolicy(ctx context.Context, t tests.T, backupPolicyID string) {
	resp, err := c.DeleteBackupPolicy(ctx, backupPolicyID)
	api.RequireNoError(t, resp, err)
}

func (c *ControlPlane) DeleteBackupPolicy(ctx context.Context, backupPolicyID string) (*http.Response, error) {
	return c.PDS.BackupPoliciesApi.ApiBackupPoliciesIdDelete(ctx, backupPolicyID).Execute()
}
