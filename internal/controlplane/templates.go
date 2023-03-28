package controlplane

import (
	"context"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
)

// Info for a single template.
type templateInfo struct {
	ID   string
	Name string
}

// Info for all app config and resource templates which belong to a data service.
type dataServiceTemplateInfo struct {
	AppConfigTemplates []templateInfo
	ResourceTemplates  []templateInfo
}

// DeleteTestStorageOptions cleans up storage options created specifically for the test run.
func (c *ControlPlane) DeleteTestStorageOptions(ctx context.Context, t tests.T) {
	resp, err := c.PDS.StorageOptionsTemplatesApi.ApiStorageOptionsTemplatesIdDelete(ctx, c.testPDSStorageTemplateID).Execute()
	api.NoErrorf(t, resp, err, "Deleting test storage options template (%s)", c.testPDSStorageTemplateID)
}

// DeleteTestApplicationTemplates cleans up application templates created specifically for the test run.
func (c *ControlPlane) DeleteTestApplicationTemplates(ctx context.Context, t tests.T) {
	for _, dsTemplate := range c.TestPDSTemplates {
		for _, configTemplateInfo := range dsTemplate.AppConfigTemplates {
			resp, err := c.PDS.ApplicationConfigurationTemplatesApi.ApiApplicationConfigurationTemplatesIdDelete(ctx, configTemplateInfo.ID).Execute()
			api.NoErrorf(t, resp, err, "Deleting configuration template (ID=%s, name=%s).", configTemplateInfo.ID, configTemplateInfo.Name)
		}

		for _, resourceTemplateInfo := range dsTemplate.ResourceTemplates {
			resp, err := c.PDS.ResourceSettingsTemplatesApi.ApiResourceSettingsTemplatesIdDelete(ctx, resourceTemplateInfo.ID).Execute()
			api.NoErrorf(t, resp, err, "Deleting resource settings template (ID=%s, name=%s)", resourceTemplateInfo.ID, resourceTemplateInfo.Name)
		}
	}
}
