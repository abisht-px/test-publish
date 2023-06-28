package controlplane

import (
	"context"
	"testing"

	apiv1 "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/assert"

	"github.com/portworx/pds-integration-test/internal/api"
)

func (s *ControlPlane) MustGetCompatibleVersions(ctx context.Context, t *testing.T) []apiv1.CompatibilityCompatibleVersions {
	response, httpResp, err := s.PDS.VersionsApi.ApiCompatibleVersionsGet(ctx).Execute()
	api.RequireNoError(t, httpResp, err)
	compatibleVersions, ok := response.GetDataOk()
	assert.Truef(t, ok, "Error getting compatible versions")
	return compatibleVersions
}
