package controlplane

import (
	"github.com/portworx/pds-integration-test/internal/api"
)

type ControlPlane struct {
	API *api.PDSClient
}

func New(apiClient *api.PDSClient) *ControlPlane {
	return &ControlPlane{
		API: apiClient,
	}
}
