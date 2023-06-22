package controlplane

import (
	"context"
	"net/http"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (c *ControlPlane) SetTestDeploymentTarget(targetID string) {
	c.testPDSDeploymentTargetID = targetID
}

func (c *ControlPlane) MustGetDeploymentTarget(ctx context.Context, t tests.T) (targetID *pds.ModelsDeploymentTarget) {
	deploymentTarget, resp, err := c.PDS.DeploymentTargetsApi.ApiDeploymentTargetsIdGet(ctx, c.testPDSDeploymentTargetID).Execute()
	api.RequireNoErrorf(t, resp, err, "Getting deployment target %s.", c.testPDSDeploymentTargetID)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return deploymentTarget
}

func (c *ControlPlane) MustWaitForDeploymentTarget(ctx context.Context, t tests.T, name string) (targetID string) {
	wait.For(t, wait.ShortTimeout, wait.RetryInterval, func(t tests.T) {
		var err error
		targetID, err = c.PDS.GetDeploymentTargetIDByName(ctx, c.TestPDSTenantID, name)
		require.NoErrorf(t, err, "PDS deployment target %q does not exist.", name)
	})

	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		err := c.PDS.CheckDeploymentTargetHealth(ctx, targetID)
		require.NoErrorf(t, err, "Deployment target %q is not healthy.", targetID)
	})
	return targetID
}

// DeleteTestDeploymentTarget deletes the default test target that was registered to the control plane.
func (s *ControlPlane) DeleteTestDeploymentTarget(ctx context.Context, t tests.T) {
	// Expect the target to be evaluated as unhealthy within 5 minutes (grace period from last received heartbeat).
	wait.For(t, 5*time.Minute, wait.RetryInterval, func(t tests.T) {
		err := s.PDS.CheckDeploymentTargetHealth(ctx, s.testPDSDeploymentTargetID)
		assert.Errorf(t, err, "Deployment target %q is still healthy.", s.testPDSDeploymentTargetID)
	})
	resp, err := s.PDS.DeploymentTargetsApi.ApiDeploymentTargetsIdDelete(ctx, s.testPDSDeploymentTargetID).Execute()
	api.NoErrorf(t, resp, err, "Deleting deployment target %s.", s.testPDSDeploymentTargetID)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Unexpected response code from deleting deployment target.")
}
