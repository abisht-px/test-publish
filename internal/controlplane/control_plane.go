package controlplane

import (
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/portworx/pds-integration-test/internal/api"
)

type ControlPlane struct {
	PDS        *api.PDSClient
	Prometheus prometheusv1.API

	TestPDSAccountID           string
	TestPDSTenantID            string
	TestPDSProjectID           string
	TestPDSNamespaceID         string
	testPDSDeploymentTargetID  string
	testPDSStorageTemplateID   string
	testPDSStorageTemplateName string
	TestPDSTemplates           map[string]dataServiceTemplateInfo
	imageVersionSpecs          []api.PDSImageReferenceSpec
}

func New(apiClient *api.PDSClient) *ControlPlane {
	return &ControlPlane{
		PDS: apiClient,
	}
}
