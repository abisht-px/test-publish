package controlplane

import (
	"context"
	"net/http"
	"testing"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	pdsDeploymentHealthStateHealthy = "Healthy"
	pdsDeploymentHealthAvailable    = "Available"
)

func (c *ControlPlane) MustDeployDeploymentSpec(ctx context.Context, t *testing.T, deployment *api.ShortDeploymentSpec) string {
	image := findImageVersionForRecord(deployment, c.imageVersionSpecs)
	require.NotNil(t, image, "No image found for deployment %s %s %s.", deployment.DataServiceName, deployment.ImageVersionTag, deployment.ImageVersionBuild)

	c.setDeploymentDefaults(deployment)

	deploymentID, err := c.PDS.CreateDeployment(ctx, deployment, image, c.TestPDSTenantID, c.testPDSDeploymentTargetID, c.TestPDSProjectID, c.testPDSNamespaceID)
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

	deployment, resp, err := s.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	if spec.ResourceSettingsTemplateName != "" {
		resourceTemplate, err := s.PDS.GetResourceSettingsTemplateByName(ctx, s.TestPDSTenantID, spec.ResourceSettingsTemplateName, *deployment.DataServiceId)
		require.NoError(t, err)
		req.ResourceSettingsTemplateId = resourceTemplate.Id
	}

	if spec.AppConfigTemplateName != "" {
		appConfigTemplate, err := s.PDS.GetAppConfigTemplateByName(ctx, s.TestPDSTenantID, spec.AppConfigTemplateName, *deployment.DataServiceId)
		require.NoError(t, err)
		req.ApplicationConfigurationTemplateId = appConfigTemplate.Id
	}

	_, resp, err = s.PDS.DeploymentsApi.ApiDeploymentsIdPut(ctx, deploymentID).Body(req).Execute()
	api.RequireNoErrorf(t, resp, err, "update %s deployment", deploymentID)
}

func (c *ControlPlane) MustWaitForDeploymentHealthy(ctx context.Context, t *testing.T, deploymentID string) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		deployment, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdStatusGet(ctx, deploymentID).Execute()
		api.RequireNoErrorf(t, resp, err, "Getting deployment %q state.", deploymentID)

		healthState := deployment.GetHealth()
		require.Equal(t, pdsDeploymentHealthStateHealthy, healthState, "Deployment %q is in state %q.", deploymentID, healthState)
	})
}

func (c *ControlPlane) MustWaitForDeploymentReplicas(ctx context.Context, t *testing.T, deploymentID string, expectedReplicas int32) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		deployment, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdStatusGet(ctx, deploymentID).Execute()
		api.RequireNoErrorf(t, resp, err, "Getting deployment %q state.", deploymentID)

		replicas := deployment.GetReplicas()
		require.Equal(t, expectedReplicas, replicas, "Deployment %q has %q replicas.", deploymentID, replicas)
	})
}

func (c *ControlPlane) MustWaitForDeploymentAvailable(ctx context.Context, t *testing.T, deploymentID string) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		deployment, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Expand("deployment_manifest").Execute()
		api.RequireNoErrorf(t, resp, err, "Getting deployment %q state.", deploymentID)

		healthState := deployment.GetDeploymentManifest().Health
		require.Equal(t, pdsDeploymentHealthAvailable, *healthState, "Deployment %q is in state %q.", deploymentID, healthState)
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

func (c *ControlPlane) MustRemoveDeployment(ctx context.Context, t *testing.T, deploymentID string) {
	resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdDelete(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)
}

func (c *ControlPlane) MustWaitForDeploymentRemoved(ctx context.Context, t *testing.T, deploymentID string) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		_, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
		assert.Errorf(t, err, "Expected an error response on getting deployment %s.", deploymentID)
		require.NotNilf(t, resp, "Received no response body while getting deployment %s.", deploymentID)
		require.Equalf(t, http.StatusNotFound, resp.StatusCode, "Deployment %s is not removed.", deploymentID)
	})
}
