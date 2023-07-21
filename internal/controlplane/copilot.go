package controlplane

import (
	"context"
	"net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"k8s.io/utils/pointer"

	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) PerformCopilotQuery(
	ctx context.Context, t tests.T, dataServiceId, query string,
) (*pds.ModelsCopilotSearchResponse, *http.Response, error) {

	requestBody := pds.RequestsCreateCopilotSearchRequest{
		Query:         pointer.String(query),
		DataServiceId: pointer.String(dataServiceId),
	}

	return c.PDS.CopilotApi.ApiCopilotSearchPost(ctx).Body(requestBody).Execute()
}
