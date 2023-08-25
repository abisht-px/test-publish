package controlplane

import (
	"context"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/pointer"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/random"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (c *ControlPlane) CreateS3BackupTarget(ctx context.Context, backupCredentialsID, bucket, region string) (*pds.ModelsBackupTarget, *http.Response, error) {
	tenantID := c.TestPDSTenantID
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	name := fmt.Sprintf("integration-test-s3-%s", nameSuffix)

	requestBody := pds.ControllersCreateTenantBackupTarget{
		Name:                &name,
		BackupCredentialsId: &backupCredentialsID,
		Bucket:              &bucket,
		Region:              &region,
		Type:                pointer.String("s3"),
	}
	return c.PDS.BackupTargetsApi.ApiTenantsIdBackupTargetsPost(ctx, tenantID).Body(requestBody).Execute()
}

func (c *ControlPlane) MustCreateS3BackupTarget(ctx context.Context, t tests.T, backupCredentialsID, bucket, region string) *pds.ModelsBackupTarget {
	backupTarget, resp, err := c.CreateS3BackupTarget(ctx, backupCredentialsID, bucket, region)
	api.RequireNoError(t, resp, err)
	return backupTarget
}

func (c *ControlPlane) MustEnsureBackupTargetCreatedInTC(ctx context.Context, t tests.T, backupTargetID string) {
	c.MustWaitForBackupTargetState(ctx, t, backupTargetID, "successful")
}

func (c *ControlPlane) MustWaitForBackupTargetState(ctx context.Context, t tests.T, backupTargetID, expectedFinalState string) {
	wait.For(t, wait.ShortTimeout, wait.ShortRetryInterval, func(t tests.T) {
		backupTargetState := c.MustGetBackupTargetState(ctx, t, backupTargetID)
		require.Equalf(t, expectedFinalState, backupTargetState.GetState(),
			"Backup target %s failed to end up in %s state to deployment target %s.", backupTargetID, expectedFinalState, c.testPDSDeploymentTargetID)
	})
}

func (c *ControlPlane) MustGetBackupTargetState(ctx context.Context, t tests.T, backupTargetID string) pds.ModelsBackupTargetState {
	backupTargetStates, resp, err := c.PDS.BackupTargetsApi.ApiBackupTargetsIdStatesGet(ctx, backupTargetID).Execute()
	api.RequireNoError(t, resp, err)

	for _, backupTargetState := range backupTargetStates.GetData() {
		if backupTargetState.GetDeploymentTargetId() == c.testPDSDeploymentTargetID {
			return backupTargetState
		}
	}
	require.Fail(t, "Backup target state for backup target %s and deployment target %s was not found.", backupTargetID, c.testPDSDeploymentTargetID)
	return pds.ModelsBackupTargetState{}
}

func (c *ControlPlane) MustDeleteBackupTarget(ctx context.Context, t tests.T, backupTargetID string) {
	// The force=true parameter ensures that all the associated backup target states are deleted even if api-workers fail
	// to delete the PX cloud credentials. This query parameter is used by default in the UI.
	resp, err := c.PDS.BackupTargetsApi.ApiBackupTargetsIdDelete(ctx, backupTargetID).Force("true").Execute()
	api.RequireNoError(t, resp, err)
	wait.For(t, wait.LongTimeout, wait.ShortRetryInterval, func(t tests.T) {
		_, resp, err := c.PDS.BackupTargetsApi.ApiBackupTargetsIdGet(ctx, backupTargetID).Execute()
		assert.Error(t, err)
		assert.NotNil(t, resp)
		require.Equalf(t, http.StatusNotFound, resp.StatusCode, "Backup target %s is not deleted.", backupTargetID)
	})
}

func (c *ControlPlane) DeleteBackupTargetIfExists(ctx context.Context, t tests.T, backupTargetID string) {
	// The force=true parameter ensures that all the associated backup target states are deleted even if api-workers fail
	// to delete the PX cloud credentials. This query parameter is used by default in the UI.
	resp, err := c.PDS.BackupTargetsApi.ApiBackupTargetsIdDelete(ctx, backupTargetID).Force("true").Execute()
	if resp.StatusCode == http.StatusNotFound {
		return
	}
	api.NoError(t, resp, err)

	wait.For(t, wait.StandardTimeout, wait.ShortRetryInterval, func(t tests.T) {
		_, resp, err := c.PDS.BackupTargetsApi.ApiBackupTargetsIdGet(ctx, backupTargetID).Execute()
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.Equalf(t, http.StatusNotFound, resp.StatusCode, "Backup target %s is not deleted.", backupTargetID)
	})
}
