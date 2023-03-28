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
