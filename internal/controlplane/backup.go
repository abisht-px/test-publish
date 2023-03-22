package controlplane

import (
	"context"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"k8s.io/utils/pointer"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) MustCreateBackup(ctx context.Context, t tests.T, deploymentID, backupTargetID string) *pds.ModelsBackup {
	requestBody := pds.ControllersCreateDeploymentBackup{
		BackupLevel:    pointer.String("snapshot"),
		BackupTargetId: pointer.String(backupTargetID),
		BackupType:     pointer.String("adhoc"),
	}
	backup, resp, err := c.API.BackupsApi.ApiDeploymentsIdBackupsPost(ctx, deploymentID).Body(requestBody).Execute()
	api.RequireNoError(t, resp, err)

	return backup
}

func (c *ControlPlane) MustDeleteBackup(ctx context.Context, t tests.T, backupID string) {
	resp, err := c.API.BackupsApi.ApiBackupsIdDelete(ctx, backupID).Execute()
	api.RequireNoError(t, resp, err)
}
