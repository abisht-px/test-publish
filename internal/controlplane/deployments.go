package controlplane

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	pdsDeploymentHealthStateHealthy = "Healthy"
	pdsDeploymentHealthAvailable    = "Available"
	pdsDeploymentHealthUnavailable  = "Unavailable"

	pdsDeploymentStateAvailable = "Available"
	pdsDeploymentStateDeploying = "Deploying"
)

func (c *ControlPlane) MustDeployDeploymentSpec(ctx context.Context, t *testing.T, deployment *api.ShortDeploymentSpec) string {
	return c.MustDeployDeploymentSpecIntoNamespace(ctx, t, deployment, c.TestPDSNamespaceID)
}

func (c *ControlPlane) MustDeployDeploymentSpecIntoNamespace(ctx context.Context, t *testing.T, deployment *api.ShortDeploymentSpec, namespaceID string) string {
	deploymentID, err := c.DeployDeploymentSpec(ctx, deployment, namespaceID)
	require.NoError(t, err, "Error while creating deployment %s.", deployment.DataServiceName)
	require.NotEmpty(t, deploymentID, "Deployment ID is empty.")

	return deploymentID
}

func (c *ControlPlane) DeployDeploymentSpec(ctx context.Context, deployment *api.ShortDeploymentSpec, namespaceID string) (string, error) {
	image := c.findImageVersionForRecord(deployment)
	if image == nil {
		return "", fmt.Errorf("no image found for deployment %s %s %s", deployment.DataServiceName, deployment.ImageVersionTag, deployment.ImageVersionBuild)
	}

	c.setDeploymentDefaults(deployment)

	return c.PDS.CreateDeployment(ctx, deployment, image, c.TestPDSTenantID, c.testPDSDeploymentTargetID, c.TestPDSProjectID, namespaceID)
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
	if (spec.BackupTargetName == "" && spec.BackupPolicyname != "") || (spec.BackupTargetName != "" && spec.BackupPolicyname == "") {
		require.FailNow(t, "backup target name and backup policy name both must be explicitly specified, and leaving either of them undefined is not allowed")
	}
	if spec.ImageVersionTag != "" || spec.ImageVersionBuild != "" {
		image := s.findImageVersionForRecord(spec)
		require.NotNil(t, image, "Update deployment: no image found for %s version.", spec.ImageVersionTag)

		req.ImageId = &image.ImageID
	}
	if spec.NodeCount != 0 {
		req.NodeCount = &spec.NodeCount
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

	if spec.BackupPolicyname != "" && spec.BackupTargetName != "" {
		backupPolicy, err := s.PDS.GetBackupPolicyByName(ctx, s.TestPDSTenantID, spec.BackupPolicyname)
		require.NoError(t, err)
		backupTarget, err := s.PDS.GetBackupTargetByName(ctx, s.TestPDSTenantID, spec.BackupTargetName)
		require.NoError(t, err)
		req.ScheduledBackup = &pds.ControllersUpdateDeploymentScheduledBackup{
			BackupPolicyId: backupPolicy.Id,
			BackupTargetId: backupTarget.Id,
		}
	}

	_, resp, err = s.PDS.DeploymentsApi.ApiDeploymentsIdPut(ctx, deploymentID).Body(req).Execute()
	api.RequireNoErrorf(t, resp, err, "update %s deployment", deploymentID)
}

func (c *ControlPlane) getDeploymentManifestHealthStatus(ctx context.Context, t tests.T, deploymentID string) (string, string) {
	deployment, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Expand("deployment_manifest").Execute()
	api.RequireNoErrorf(t, resp, err, "Getting deployment %q state.", deploymentID)

	manifest := deployment.GetDeploymentManifest()
	return *manifest.Health, *manifest.Status
}

func (c *ControlPlane) MustWaitForDeploymentManifestInitialChange(ctx context.Context, t *testing.T, deploymentID string) {
	wait.For(t, wait.StandardTimeout, wait.ShortRetryInterval, func(t tests.T) {
		health, status := c.getDeploymentManifestHealthStatus(ctx, t, deploymentID)
		require.NotEqual(t, pdsDeploymentHealthUnavailable, health, "Deployment %q has health %q.", deploymentID, health)
		require.NotEqual(t, pdsDeploymentStateDeploying, status, "Deployment %q is in state %q.", deploymentID, status)
	})
}

func (c *ControlPlane) MustDeploymentManifestStatusHealthAvailable(ctx context.Context, t *testing.T, deploymentID string) {
	health, status := c.getDeploymentManifestHealthStatus(ctx, t, deploymentID)
	require.Equal(t, pdsDeploymentHealthAvailable, health, "Deployment %q has health %q.", deploymentID, health)
	require.Equal(t, pdsDeploymentStateAvailable, status, "Deployment %q is in state %q.", deploymentID, status)
}

func (c *ControlPlane) MustDeploymentManifestStatusHealthUnavailable(ctx context.Context, t *testing.T, deploymentID string) {
	health, status := c.getDeploymentManifestHealthStatus(ctx, t, deploymentID)
	require.Equal(t, pdsDeploymentHealthUnavailable, health, "Deployment %q has health %q.", deploymentID, health)
	require.Equal(t, pdsDeploymentStateDeploying, status, "Deployment %q is in state %q.", deploymentID, status)
}

func (c *ControlPlane) MustWaitForDeploymentHealthy(ctx context.Context, t *testing.T, deploymentID string) {
	deployment, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	wait.For(t, dataservices.GetLongTimeoutFor(*deployment.NodeCount), wait.RetryInterval, func(t tests.T) {
		deployment, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdStatusGet(ctx, deploymentID).Execute()
		api.RequireNoErrorf(t, resp, err, "Getting deployment %q state.", deploymentID)

		healthState := deployment.GetHealth()
		require.Equal(t, pdsDeploymentHealthStateHealthy, healthState, "Deployment %q is in state %q.", deploymentID, healthState)
	})
}

func (c *ControlPlane) GetDeploymentById(ctx context.Context, t *testing.T, deploymentID string) (*pds.ModelsDeployment, *http.Response, error) {
	return c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
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
	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		deployment, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Expand("deployment_manifest").Execute()
		api.RequireNoErrorf(t, resp, err, "Getting deployment %q state.", deploymentID)

		healthState := deployment.GetDeploymentManifest().Health
		require.Equal(t, pdsDeploymentHealthAvailable, *healthState, "Deployment %q is in state %q.", deploymentID, healthState)
	})
}

func (c *ControlPlane) MustWaitForDeploymentEventCondition(
	ctx context.Context, t *testing.T,
	deploymentID string,
	eventPredicate func(event pds.ModelsDeploymentTargetDeploymentEvent) bool,
	description string,
) {
	wait.For(t, wait.ShortTimeout, wait.RetryInterval, func(t tests.T) {
		eventsResponse, resp, err := c.PDS.EventsApi.ApiDeploymentsIdEventsGet(ctx, deploymentID).Execute()
		api.RequireNoErrorf(t, resp, err, "Getting deployment %q events.", deploymentID)

		hasEvent := hasMatchingEvent(eventsResponse, eventPredicate)
		require.Truef(t, hasEvent, "No event matches condition: %s.", description)
	})
}

func hasMatchingEvent(events []pds.ModelsDeploymentTargetDeploymentEvent, predicate func(event pds.ModelsDeploymentTargetDeploymentEvent) bool) bool {
	for _, resourceEvent := range events {
		if predicate(resourceEvent) {
			return true
		}
	}
	return false
}

func (c *ControlPlane) findImageVersionForRecord(deployment *api.ShortDeploymentSpec) *api.PDSImageReferenceSpec {
	for _, image := range c.imageVersionSpecs {
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

// SetDefaultImageVersionBuild sets the latest build tag for the deployment (if not set yet or is forced to overwrite).
func (c *ControlPlane) SetDefaultImageVersionBuild(deployment *api.ShortDeploymentSpec, overwrite bool) {
	if deployment.ImageVersionBuild == "" || overwrite {
		if overwrite {
			deployment.ImageVersionBuild = ""
		}
		image := c.findImageVersionForRecord(deployment)
		if image != nil {
			deployment.ImageVersionBuild = image.ImageVersionBuild
		}
	}
}

func (c *ControlPlane) MustRemoveDeployment(ctx context.Context, t *testing.T, deploymentID string) {
	resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdDelete(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)
}

func (c *ControlPlane) MustRemoveDeploymentIfExists(ctx context.Context, t *testing.T, deploymentID string) {
	_, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	if err == nil || resp == nil || resp.StatusCode != http.StatusNotFound {
		c.MustRemoveDeployment(ctx, t, deploymentID)
	}
}

func (c *ControlPlane) MustWaitForDeploymentRemoved(ctx context.Context, t *testing.T, deploymentID string) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		_, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
		assert.Errorf(t, err, "Expected an error response on getting deployment %s.", deploymentID)
		require.NotNilf(t, resp, "Received no response body while getting deployment %s.", deploymentID)
		require.Equalf(t, http.StatusNotFound, resp.StatusCode, "Deployment %s is not removed.", deploymentID)
	})
}

func (c *ControlPlane) MustHaveDeploymentEventsSorted(ctx context.Context, t *testing.T, deploymentID string) {
	eventsResponse, resp, err := c.PDS.EventsApi.ApiDeploymentsIdEventsGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	n := len(eventsResponse)
	for i := 1; i < n; i++ {
		x, err := time.Parse(time.RFC3339, *(eventsResponse[i-1].Timestamp))
		assert.NoError(t, err, "Error while parsing time %v", *(eventsResponse[i-1].Timestamp))

		y, err := time.Parse(time.RFC3339, *(eventsResponse[i].Timestamp))
		assert.NoError(t, err, "Error while parsing time %v", *(eventsResponse[i].Timestamp))

		assert.Truef(t, x.After(y) || x.Equal(y), "Events are not sorted based on timestamp")
	}
}

func (c *ControlPlane) MustHaveNoDuplicateDeploymentEvents(ctx context.Context, t *testing.T, deploymentID string) {
	eventsResponse, resp, err := c.PDS.EventsApi.ApiDeploymentsIdEventsGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	m := make(map[string]bool)
	for _, e := range eventsResponse {
		if _, ok := m[*e.Name]; ok {
			assert.Truef(t, !ok, "Duplicate event %s found", *e.Name)
		} else {
			m[*e.Name] = true
		}
	}
}

func (c *ControlPlane) MustGetErrorOnDeploymentEventsGet(ctx context.Context, t *testing.T, deploymentID string) {
	_, _, err := c.PDS.EventsApi.ApiDeploymentsIdEventsGet(ctx, deploymentID).Execute()
	assert.Errorf(t, err, "Expected an error response on getting deployment events for deployment %s.", deploymentID)
}

func (s *ControlPlane) FailUpdateDeployment(ctx context.Context, t *testing.T, deploymentID string, spec *api.ShortDeploymentSpec) {
	req := pds.ControllersUpdateDeploymentRequest{}
	nodeCount := int32(10)
	if spec.ImageVersionTag != "" || spec.ImageVersionBuild != "" {
		image := s.findImageVersionForRecord(spec)
		require.NotNil(t, image, "Update deployment: no image found for %s version.", spec.ImageVersionTag)

		req.ImageId = &image.ImageID
	}
	if spec.NodeCount != 0 {
		req.NodeCount = &spec.NodeCount
	} else {
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
	api.RequireErrorWithStatus(t, resp, err, http.StatusBadRequest)
}

func (c *ControlPlane) MustHaveDeploymentEventsForCorrectDeployment(ctx context.Context, t *testing.T, deploymentID string) {
	deployment, resp, err := c.PDS.DeploymentsApi.ApiDeploymentsIdGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	eventsResponse, resp, err := c.PDS.EventsApi.ApiDeploymentsIdEventsGet(ctx, deploymentID).Execute()
	api.RequireNoError(t, resp, err)

	n := len(eventsResponse)
	for i := 1; i < n; i++ {
		assert.Contains(t, *eventsResponse[i].ResourceName, *deployment.ClusterResourceName, "Resource name does not contain deployment name. expected %s in resource name, got %s", deployment.ClusterResourceName, *eventsResponse[i].ResourceName)
	}
}

func (c *ControlPlane) ListDeploymentsInProject(ctx context.Context, projectID string) ([]pds.ModelsDeployment, error) {
	getDeploymentsReq := c.PDS.DeploymentsApi.ApiProjectsIdDeploymentsGet(ctx, projectID)

	deployments, resp, err := getDeploymentsReq.Execute()
	if err != nil {
		return nil, api.ExtractErrorDetails(resp, err)
	}

	return deployments.GetData(), nil
}

func (c *ControlPlane) ListDeploymentsForDeploymentTarget(
	ctx context.Context,
	projectID string,
	deploymentTargetID string,
) ([]pds.ModelsDeployment, error) {
	getDeploymentsReq := c.PDS.DeploymentsApi.ApiProjectsIdDeploymentsGet(ctx, projectID).
		DeploymentTargetId(deploymentTargetID)

	deployments, resp, err := getDeploymentsReq.Execute()
	if err != nil {
		return nil, api.ExtractErrorDetails(resp, err)
	}

	return deployments.GetData(), nil
}
