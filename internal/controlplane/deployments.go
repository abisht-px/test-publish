package controlplane

import (
	"context"
	"testing"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	pdsDeploymentHealthStateHealthy = "Healthy"
)

func (c *ControlPlane) MustDeployDeploymentSpec(ctx context.Context, t *testing.T, deployment *api.ShortDeploymentSpec) string {
	image := findImageVersionForRecord(deployment, c.imageVersionSpecs)
	require.NotNil(t, image, "No image found for deployment %s %s %s.", deployment.DataServiceName, deployment.ImageVersionTag, deployment.ImageVersionBuild)

	c.setDeploymentDefaults(deployment)

	deploymentID, err := c.API.CreateDeployment(ctx, deployment, image, c.TestPDSTenantID, c.testPDSDeploymentTargetID, c.TestPDSProjectID, c.testPDSNamespaceID)
	require.NoError(t, err, "Error while creating deployment %s.", deployment.DataServiceName)
	require.NotEmpty(t, deploymentID, "Deployment ID is empty.")

	return deploymentID
}

func (c *ControlPlane) setDeploymentDefaults(deployment *api.ShortDeploymentSpec) {
	if deployment.ServiceType == "" {
		deployment.ServiceType = "ClusterIP"
	}
	if deployment.StorageOptionName == "" {
		deployment.StorageOptionName = c.testPDSStorageTemplateName
	}
	dsTemplates, found := c.TestPDSTemplates[deployment.DataServiceName]
	if found {
		if deployment.ResourceSettingsTemplateName == "" {
			deployment.ResourceSettingsTemplateName = dsTemplates.ResourceTemplates[0].Name
		}
		if deployment.AppConfigTemplateName == "" {
			deployment.AppConfigTemplateName = dsTemplates.AppConfigTemplates[0].Name
		}
	}
}

func (s *ControlPlane) MustUpdateDeployment(ctx context.Context, t *testing.T, deploymentID string, spec *api.ShortDeploymentSpec) {
	req := pds.ControllersUpdateDeploymentRequest{}
	if spec.ImageVersionTag != "" || spec.ImageVersionBuild != "" {
		image := findImageVersionForRecord(spec, s.imageVersionSpecs)
		require.NotNil(t, image, "Update deployment: no image found for %s version.", spec.ImageVersionTag)

		req.ImageId = &image.ImageID
	}
	if spec.NodeCount != 0 {
		nodeCount := int32(spec.NodeCount)
		req.NodeCount = &nodeCount
	}

	deployment, resp, err := s.API.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	if spec.ResourceSettingsTemplateName != "" {
		resourceTemplate, err := s.API.GetResourceSettingsTemplateByName(ctx, s.TestPDSTenantID, spec.ResourceSettingsTemplateName, *deployment.DataServiceId)
		require.NoError(t, err)
		req.ResourceSettingsTemplateId = resourceTemplate.Id
	}

	if spec.AppConfigTemplateName != "" {
		appConfigTemplate, err := s.API.GetAppConfigTemplateByName(ctx, s.TestPDSTenantID, spec.AppConfigTemplateName, *deployment.DataServiceId)
		require.NoError(t, err)
		req.ApplicationConfigurationTemplateId = appConfigTemplate.Id
	}

	_, resp, err = s.API.DeploymentsApi.ApiDeploymentsIdPut(ctx, deploymentID).Body(req).Execute()
	api.RequireNoErrorf(t, resp, err, "update %s deployment", deploymentID)
}

func (c *ControlPlane) MustWaitForDeploymentHealthy(ctx context.Context, t *testing.T, deploymentID string) {
	wait.For(t, wait.DeploymentStatusHealthyTimeout, wait.RetryInterval, func(t tests.T) {
		deployment, resp, err := c.API.DeploymentsApi.ApiDeploymentsIdStatusGet(ctx, deploymentID).Execute()
		err = api.ExtractErrorDetails(resp, err)
		require.NoError(t, err, "Getting deployment %q state.", deploymentID)

		healthState := deployment.GetHealth()
		require.Equal(t, pdsDeploymentHealthStateHealthy, healthState, "Deployment %q is in state %q.", deploymentID, healthState)
	})
}

func findImageVersionForRecord(deployment *api.ShortDeploymentSpec, images []api.PDSImageReferenceSpec) *api.PDSImageReferenceSpec {
	for _, image := range images {
		found := image.DataServiceName == deployment.DataServiceName
		if deployment.ImageVersionTag != "" {
			found = found && image.ImageVersionTag == deployment.ImageVersionTag
		}
		if deployment.ImageVersionBuild != "" {
			found = found && image.ImageVersionBuild == deployment.ImageVersionBuild
		}
		if found {
			return &image
		}
	}
	return nil
}
