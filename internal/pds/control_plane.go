package pds

import (
	"github.com/portworx/pds-integration-test/internal/api"
)

type ControlPlane struct {
	API *api.PDSClient
}

func NewControlPlane(apiClient *api.PDSClient) *ControlPlane {
	return &ControlPlane{
		API: apiClient,
	}
}
