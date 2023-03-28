package controlplane

import (
	"time"

	"github.com/portworx/pds-integration-test/internal/api"
)

const (
	waiterRetryInterval      = time.Second * 10
	waiterShortRetryInterval = time.Second * 1
)

type ControlPlane struct {
	API *api.PDSClient

	// TODO 4915: Unexport fields when all referencing helpers are attached to the controlPlane.
	TestPDSAccountID           string
	TestPDSTenantID            string
	TestPDSProjectID           string
	TestPDSNamespaceID         string
	testPDSDeploymentTargetID  string
	TestPDSStorageTemplateID   string
	TestPDSStorageTemplateName string
	TestPDSTemplatesMap        map[string]dataServiceTemplateInfo
	ImageVersionSpecList       []api.PDSImageReferenceSpec
}

func New(apiClient *api.PDSClient) *ControlPlane {
	return &ControlPlane{
		API: apiClient,
	}
}
