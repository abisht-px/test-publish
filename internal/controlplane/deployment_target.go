package controlplane

import (
	"context"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	waiterDeploymentTargetStatusUnhealthyTimeout = time.Minute * 5
	waiterDeploymentTargetNameExistsTimeout      = time.Second * 90
	waiterDeploymentTargetStatusHealthyTimeout   = time.Minute * 10
)

func (c *ControlPlane) SetTestDeploymentTarget(targetID string) {
	c.testPDSDeploymentTargetID = targetID
}

func (c *ControlPlane) MustWaitForDeploymentTarget(ctx context.Context, t tests.T, name string) (targetID string) {
	wait.For(t, waiterDeploymentTargetNameExistsTimeout, waiterRetryInterval, func(t tests.T) {
		var err error
		targetID, err = c.API.GetDeploymentTargetIDByName(ctx, c.TestPDSTenantID, name)
		require.NoErrorf(t, err, "PDS deployment target %q does not exist.", name)
	})

	wait.For(t, waiterDeploymentTargetStatusHealthyTimeout, waiterRetryInterval, func(t tests.T) {
		err := c.API.CheckDeploymentTargetHealth(ctx, targetID)
		require.NoErrorf(t, err, "Deployment target %q is not healthy.", targetID)
	})
	return targetID
}

// DeleteTestDeploymentTarget deletes the default test target that was registered to the control plane.
func (s *ControlPlane) DeleteTestDeploymentTarget(ctx context.Context, t tests.T) {
	wait.For(t, waiterDeploymentTargetStatusUnhealthyTimeout, waiterRetryInterval, func(t tests.T) {
		err := s.API.CheckDeploymentTargetHealth(ctx, s.testPDSDeploymentTargetID)
		assert.Errorf(t, err, "Deployment target %q is still healthy.", s.testPDSDeploymentTargetID)
	})
	resp, err := s.API.DeploymentTargetsApi.ApiDeploymentTargetsIdDelete(ctx, s.testPDSDeploymentTargetID).Execute()
	api.NoErrorf(t, resp, err, "Deleting deployment target %s.", s.testPDSDeploymentTargetID)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Unexpected response code from deleting deployment target.")
}
