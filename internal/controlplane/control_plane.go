package controlplane

import (
	"github.com/portworx/pds-integration-test/internal/api"
)

type ControlPlane struct {
	API *api.PDSClient

	testPDSAccountID           string
	TestPDSTenantID            string
	TestPDSProjectID           string
	testPDSNamespaceID         string
	testPDSDeploymentTargetID  string
	testPDSStorageTemplateID   string
	testPDSStorageTemplateName string
	TestPDSTemplates           map[string]dataServiceTemplateInfo
	imageVersionSpecs          []api.PDSImageReferenceSpec
}

func New(apiClient *api.PDSClient) *ControlPlane {
	return &ControlPlane{
		API: apiClient,
	}
}
