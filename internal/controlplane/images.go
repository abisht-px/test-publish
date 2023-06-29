package controlplane

import (
	"context"
	"testing"

	apiv1 "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/assert"

	"github.com/portworx/pds-integration-test/internal/api"
)

func (c *ControlPlane) MustGetAllImagesForDataService(ctx context.Context, t *testing.T, dataServiceID string) []apiv1.ModelsImage {
	response, httpResp, err := c.PDS.ImagesApi.ApiImagesGet(ctx).DataServiceId(dataServiceID).SortBy("-created_at").Latest(false).Limit("1000").Execute()
	api.RequireNoError(t, httpResp, err)
	images, ok := response.GetDataOk()
	assert.Truef(t, ok, "Error getting images")
	assert.NotEmptyf(t, images, "No images returned for data service ID %s", dataServiceID)
	return images
}
