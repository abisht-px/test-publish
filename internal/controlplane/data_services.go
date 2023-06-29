package controlplane

import (
	"context"
	"testing"

	apiv1 "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/assert"

	"github.com/portworx/pds-integration-test/internal/api"
)

func (c *ControlPlane) MustGetDataServicesByName(ctx context.Context, t *testing.T) map[string]apiv1.ModelsDataService {
	response, httpResp, err := c.PDS.DataServicesApi.ApiDataServicesGet(ctx).Execute()
	api.RequireNoError(t, httpResp, err)
	dataServices, ok := response.GetDataOk()
	assert.Truef(t, ok, "Error getting data services")
	assert.NotEmptyf(t, dataServices, "No data services returned")

	dataServicesByName := make(map[string]apiv1.ModelsDataService)
	for i := range dataServices {
		dataService := dataServices[i]
		dataServicesByName[*dataService.Name] = dataService
	}

	return dataServicesByName
}
