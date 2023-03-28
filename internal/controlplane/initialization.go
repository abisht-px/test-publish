package controlplane

import (
	"context"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/pointer"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func (c *ControlPlane) MustInitializeTestData(
	ctx context.Context, t tests.T,
	accountName, tenantName, projectName, namePrefix string,
) {
	c.mustHavePDStestAccount(ctx, t, accountName)
	c.mustHavePDStestTenant(ctx, t, tenantName)
	c.mustHavePDStestProject(ctx, t, projectName)
	c.mustLoadImageVersions(ctx, t)
	c.mustCreateStorageOptions(ctx, t, namePrefix)
	c.mustCreateApplicationTemplates(ctx, t, namePrefix)
}

func (c *ControlPlane) mustHavePDStestAccount(ctx context.Context, t tests.T, name string) {
	// TODO: Use account name query filters
	accounts, resp, err := c.API.AccountsApi.ApiAccountsGet(ctx).Execute()
	api.RequireNoError(t, resp, err)
	require.NotEmpty(t, accounts, "PDS API must return at least one account.")

	var testPDSAccountID string
	for _, account := range accounts.GetData() {
		if account.GetName() == name {
			testPDSAccountID = account.GetId()
			break
		}
	}
	require.NotEmpty(t, testPDSAccountID, "PDS account %s not found.", name)
	c.testPDSAccountID = testPDSAccountID
}

func (c *ControlPlane) mustHavePDStestTenant(ctx context.Context, t tests.T, name string) {
	// TODO: Use tenant name query filters
	tenants, resp, err := c.API.TenantsApi.ApiAccountsIdTenantsGet(ctx, c.testPDSAccountID).Execute()
	api.RequireNoError(t, resp, err)
	require.NotEmpty(t, tenants, "PDS API must return at least one tenant.")

	var testPDSTenantID string
	for _, tenant := range tenants.GetData() {
		if tenant.GetName() == name {
			testPDSTenantID = tenant.GetId()
			break
		}
	}
	require.NotEmpty(t, testPDSTenantID, "PDS tenant %s not found.", name)
	c.TestPDSTenantID = testPDSTenantID
}

func (c *ControlPlane) mustHavePDStestProject(ctx context.Context, t tests.T, name string) {
	// TODO: Use project name query filters
	projects, resp, err := c.API.ProjectsApi.ApiTenantsIdProjectsGet(ctx, c.TestPDSTenantID).Execute()
	api.RequireNoError(t, resp, err)
	require.NotEmpty(t, projects, "PDS API must return at least one project.")

	var testPDSProjectID string
	for _, project := range projects.GetData() {
		if project.GetName() == name {
			testPDSProjectID = project.GetId()
			break
		}
	}
	require.NotEmpty(t, testPDSProjectID, "PDS project %s not found.", name)
	c.TestPDSProjectID = testPDSProjectID
}

func (c *ControlPlane) mustLoadImageVersions(ctx context.Context, t tests.T) {
	imageVersions, err := c.API.GetAllImageVersions(ctx)
	require.NoError(t, err, "Error while reading image versions.")
	require.NotEmpty(t, imageVersions, "No image versions found.")
	c.imageVersionSpecs = imageVersions
}

func (c *ControlPlane) mustCreateStorageOptions(ctx context.Context, t tests.T, namePrefix string) {
	storageTemplate := pds.ControllersCreateStorageOptionsTemplateRequest{
		Name:   pointer.StringPtr(namePrefix),
		Repl:   pointer.Int32Ptr(1),
		Secure: pointer.BoolPtr(false),
		Fs:     pointer.StringPtr("xfs"),
		Fg:     pointer.BoolPtr(false),
	}
	storageTemplateResp, resp, err := c.API.StorageOptionsTemplatesApi.
		ApiTenantsIdStorageOptionsTemplatesPost(ctx, c.TestPDSTenantID).
		Body(storageTemplate).Execute()
	api.RequireNoError(t, resp, err)
	require.NoError(t, err)

	c.testPDSStorageTemplateID = storageTemplateResp.GetId()
	c.testPDSStorageTemplateName = storageTemplateResp.GetName()
}

func (c *ControlPlane) mustCreateApplicationTemplates(ctx context.Context, t tests.T, namePrefix string) {
	dataServicesTemplates := make(map[string]dataServiceTemplateInfo)
	for _, imageVersion := range c.imageVersionSpecs {
		templatesSpec, found := dataservices.TemplateSpecs[imageVersion.DataServiceName]
		if !found {
			continue
		}
		_, found = dataServicesTemplates[imageVersion.DataServiceName]
		if found {
			continue
		}

		var resultTemplateInfo dataServiceTemplateInfo
		for _, configTemplateSpec := range templatesSpec.ConfigurationTemplates {
			configTemplateBody := configTemplateSpec
			if configTemplateBody.Name == nil {
				configTemplateBody.Name = pointer.StringPtr(namePrefix)
			}
			configTemplateBody.DataServiceId = pds.PtrString(imageVersion.DataServiceID)

			configTemplate, resp, err := c.API.ApplicationConfigurationTemplatesApi.
				ApiTenantsIdApplicationConfigurationTemplatesPost(ctx, c.TestPDSTenantID).
				Body(configTemplateBody).Execute()
			api.RequireNoError(t, resp, err)

			configTemplateInfo := templateInfo{
				ID:   configTemplate.GetId(),
				Name: configTemplate.GetName(),
			}

			resultTemplateInfo.AppConfigTemplates = append(resultTemplateInfo.AppConfigTemplates, configTemplateInfo)
		}

		for _, resourceTemplateSpec := range templatesSpec.ResourceTemplates {
			resourceTemplateBody := resourceTemplateSpec
			if resourceTemplateBody.Name == nil {
				resourceTemplateBody.Name = pointer.StringPtr(namePrefix)
			}
			resourceTemplateBody.DataServiceId = pds.PtrString(imageVersion.DataServiceID)

			resourceTemplate, resp, err := c.API.ResourceSettingsTemplatesApi.
				ApiTenantsIdResourceSettingsTemplatesPost(ctx, c.TestPDSTenantID).
				Body(resourceTemplateBody).Execute()
			api.RequireNoError(t, resp, err)

			resourceTemplateInfo := templateInfo{
				ID:   resourceTemplate.GetId(),
				Name: resourceTemplate.GetName(),
			}

			resultTemplateInfo.ResourceTemplates = append(resultTemplateInfo.ResourceTemplates, resourceTemplateInfo)
		}

		dataServicesTemplates[imageVersion.DataServiceName] = resultTemplateInfo
	}
	c.TestPDSTemplates = dataServicesTemplates
}
