package controlplane

import (
	"context"
	"net/http"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	waiterDeploymentTargetStatusUnhealthyTimeout = time.Minute * 5
)

// DeleteTestDeploymentTarget deletes the default test target that was registered to the control plane.
func (s *ControlPlane) DeleteTestDeploymentTarget(ctx context.Context, t tests.T) {
	wait.For(t, waiterDeploymentTargetStatusUnhealthyTimeout, waiterRetryInterval, func(t tests.T) {
		err := s.API.CheckDeploymentTargetHealth(ctx, s.TestPDSDeploymentTargetID)
		assert.Errorf(t, err, "Deployment target %q is still healthy.", s.TestPDSDeploymentTargetID)
	})
	resp, err := s.API.DeploymentTargetsApi.ApiDeploymentTargetsIdDelete(ctx, s.TestPDSDeploymentTargetID).Execute()
	api.NoErrorf(t, resp, err, "Deleting deployment target %s.", s.TestPDSDeploymentTargetID)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Unexpected response code from deleting deployment target.")
}
