package controlplane

import (
	"github.com/portworx/pds-integration-test/internal/api"
)

type ControlPlane struct {
	API *api.PDSClient

	// TODO 4915: Unexport fields when all referencing helpers are attached to the controlPlane.
	TestPDSAccountID           string
	TestPDSTenantID            string
	TestPDSProjectID           string
	TestPDSNamespaceID         string
	TestPDSDeploymentTargetID  string
	TestPDSServiceAccountID    string
	TestPDSStorageTemplateID   string
	TestPDSStorageTemplateName string
	TestPDSTemplatesMap        map[string]DataServiceTemplateInfo
	ImageVersionSpecList       []api.PDSImageReferenceSpec
}

func New(apiClient *api.PDSClient) *ControlPlane {
	return &ControlPlane{
		API: apiClient,
	}
}
