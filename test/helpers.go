package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
)

const (
	pdsDeploymentTargetHealthState = "healthy"
	pdsDeploymentHealthState       = "Healthy"
)

func isDeploymentTargetHealthy(t *testing.T, ctx context.Context, apiClient *pds.APIClient, deploymentTargetID string) bool {
	target, resp, err := apiClient.DeploymentTargetsApi.ApiDeploymentTargetsIdGet(ctx, deploymentTargetID).Execute()
	requireNoAPIError(t, resp, err)
	return target.GetStatus() == pdsDeploymentTargetHealthState
}

func getDeploymentTargetIDByName(t *testing.T, ctx context.Context, apiClient *pds.APIClient, tenantID, deploymentTargetName string) (string, error) {
	targets, resp, err := apiClient.DeploymentTargetsApi.ApiTenantsIdDeploymentTargetsGet(ctx, tenantID).Execute()
	requireNoAPIError(t, resp, err)
	for _, target := range targets.GetData() {
		if target.GetName() == deploymentTargetName {
			return target.GetId(), nil
		}
	}
	return "", fmt.Errorf("deployment target %s not found", deploymentTargetName)
}

func getNamespaceIDByName(t *testing.T, ctx context.Context, apiClient *pds.APIClient, deploymentTargetID, namespaceName string) (string, error) {
	namespaces, resp, err := apiClient.NamespacesApi.ApiDeploymentTargetsIdNamespacesGet(ctx, deploymentTargetID).Execute()
	requireNoAPIError(t, resp, err)
	for _, namespace := range namespaces.GetData() {
		if namespace.GetName() == namespaceName {
			return namespace.GetId(), nil
		}
	}
	return "", fmt.Errorf("namespace %s not found", namespaceName)
}

func isDeploymentHealthy(t *testing.T, ctx context.Context, apiClient *pds.APIClient, deploymentID string) bool {
	deployment, resp, err := apiClient.DeploymentsApi.ApiDeploymentsIdStatusGet(ctx, deploymentID).Execute()
	requireNoAPIError(t, resp, err)
	return deployment.GetHealth() == pdsDeploymentHealthState
}

func getAllImageVersions(t *testing.T, ctx context.Context, apiClient *pds.APIClient) ([]PDSImageReferenceSpec, error) {
	var records []PDSImageReferenceSpec

	dataServices, resp, err := apiClient.DataServicesApi.ApiDataServicesGet(ctx).Execute()
	requireNoAPIError(t, resp, err)

	dataServicesByID := make(map[string]pds.ModelsDataService)
	for i := range dataServices.GetData() {
		dataService := dataServices.GetData()[i]
		dataServicesByID[dataService.GetId()] = dataService
	}

	images, resp, err := apiClient.ImagesApi.ApiImagesGet(ctx).Execute()
	requireNoAPIError(t, resp, err)

	for _, image := range images.GetData() {
		dataService := dataServicesByID[image.GetDataServiceId()]
		record := PDSImageReferenceSpec{
			DataServiceName:   dataService.GetName(),
			DataServiceID:     dataService.GetId(),
			VersionID:         image.GetVersionId(),
			ImageVersionBuild: image.GetBuild(),
			ImageVersionTag:   image.GetTag(),
			ImageID:           image.GetId(),
		}
		records = append(records, record)
	}

	return records, nil
}

func findImageVersionForRecord(deployment *ShortDeploymentSpec, images []PDSImageReferenceSpec) *PDSImageReferenceSpec {
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

func getResourceSettingsTemplateByName(t *testing.T, ctx context.Context, apiClient *pds.APIClient, tenantID, templateName, dataServiceID string) (*pds.ModelsResourceSettingsTemplate, error) {
	resources, resp, err := apiClient.ResourceSettingsTemplatesApi.ApiTenantsIdResourceSettingsTemplatesGet(ctx, tenantID).Name(templateName).Execute()
	requireNoAPIError(t, resp, err)
	for _, r := range resources.GetData() {
		if r.GetDataServiceId() == dataServiceID {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("resource settings template %s not found", templateName)
}

func getAppConfigTemplateByName(t *testing.T, ctx context.Context, apiClient *pds.APIClient, tenantID, templateName, dataServiceID string) (*pds.ModelsApplicationConfigurationTemplate, error) {
	appConfigurations, resp, err := apiClient.ApplicationConfigurationTemplatesApi.ApiTenantsIdApplicationConfigurationTemplatesGet(ctx, tenantID).Name(templateName).Execute()
	requireNoAPIError(t, resp, err)
	for _, c := range appConfigurations.GetData() {
		if c.GetDataServiceId() == dataServiceID {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("application configuration template %s not found", templateName)
}

func createPDSDeployment(t *testing.T, ctx context.Context, apiClient *pds.APIClient, deployment *ShortDeploymentSpec, image *PDSImageReferenceSpec, tenantID, deploymentTargetID, projectID, namespaceID string) (string, error) {
	resource, err := getResourceSettingsTemplateByName(t, ctx, apiClient, tenantID, deployment.ResourceSettingsTemplateName, image.DataServiceID)
	if err != nil {
		return "", err
	}

	appConfig, err := getAppConfigTemplateByName(t, ctx, apiClient, tenantID, deployment.AppConfigTemplateName, image.DataServiceID)
	if err != nil {
		return "", err
	}

	storages, resp, err := apiClient.StorageOptionsTemplatesApi.ApiTenantsIdStorageOptionsTemplatesGet(ctx, tenantID).Name(deployment.StorageOptionName).Execute()
	requireNoAPIError(t, resp, err)

	if len(storages.GetData()) == 0 {
		return "", fmt.Errorf("storage option template %s not found", deployment.StorageOptionName)
	}
	if len(storages.GetData()) != 1 {
		return "", fmt.Errorf("more than one storage option template found")
	}
	storage := storages.GetData()[0]

	var backupPolicy *pds.ModelsBackupPolicy
	if len(deployment.BackupPolicyname) > 0 {
		backupPolicies, resp, err := apiClient.BackupPoliciesApi.ApiTenantsIdBackupPoliciesGet(ctx, tenantID).Name(deployment.BackupPolicyname).Execute()
		requireNoAPIError(t, resp, err)
		if len(backupPolicies.GetData()) == 0 {
			return "", fmt.Errorf("backup policy %s not found", deployment.BackupPolicyname)
		}
		if len(backupPolicies.GetData()) != 1 {
			return "", fmt.Errorf("more than one backup policy found")
		}
		backupPolicy = &backupPolicies.GetData()[0]
	}

	dns, resp, err := apiClient.TenantsApi.ApiTenantsIdDnsDetailsGet(ctx, tenantID).Execute()
	requireNoAPIError(t, resp, err)

	pdsDeployment := pds.NewControllersCreateProjectDeployment()
	pdsDeployment.SetApplicationConfigurationTemplateId(appConfig.GetId())
	pdsDeployment.SetDeploymentTargetId(deploymentTargetID)
	pdsDeployment.SetDnsZone(dns.GetDnsZone())
	pdsDeployment.SetImageId(image.ImageID)
	pdsDeployment.SetName(deployment.NamePrefix)
	pdsDeployment.SetNamespaceId(namespaceID)
	pdsDeployment.SetNodeCount(int32(deployment.NodeCount))
	pdsDeployment.SetResourceSettingsTemplateId(resource.GetId())
	if backupPolicy != nil {
		pdsDeployment.ScheduledBackup.SetBackupPolicyId(backupPolicy.GetId())
	}
	pdsDeployment.SetServiceType(deployment.ServiceType)
	pdsDeployment.SetStorageOptionsTemplateId(storage.GetId())

	res, httpRes, err := apiClient.DeploymentsApi.ApiProjectsIdDeploymentsPost(ctx, projectID).Body(*pdsDeployment).Execute()
	if err != nil {
		rawbody, parseErr := io.ReadAll(httpRes.Body)
		if parseErr != nil {
			return "", err
		}
		return "", fmt.Errorf(string(rawbody))
	}

	return res.GetId(), nil
}

func isJobSucceeded(job *batchv1.Job) bool {
	return *job.Spec.Completions == job.Status.Succeeded
}

func isBackupFinished(backup *backupsv1.Backup) bool {
	return isBackupSucceeded(backup) || isBackupFailed(backup)
}

func isBackupSucceeded(backup *backupsv1.Backup) bool {
	return backup.Status.Succeeded > 0
}

func isBackupFailed(backup *backupsv1.Backup) bool {
	return backup.Status.Failed > 0
}

func requireNoAPIError(t *testing.T, resp *http.Response, err error) {
	t.Helper()

	if err != nil {
		rawbody, parseErr := io.ReadAll(resp.Body)
		require.NoError(t, parseErr, "Error calling PDS API: failed to read response body")
		require.NoErrorf(t, err, "Error calling PDS API: %s", rawbody)
	}
}

func requireNoAPIErrorf(t *testing.T, resp *http.Response, err error, msg string, args ...any) {
	t.Helper()

	if err != nil {
		rawbody, parseErr := io.ReadAll(resp.Body)
		require.NoError(t, parseErr, "Error calling PDS API: failed to read response body")
		details := fmt.Sprintf(msg, args...)
		require.NoErrorf(t, err, "Error calling PDS API: %s: %s", details, rawbody)
	}
}

type TestLogger struct {
	t *testing.T
}

func (l *TestLogger) Print(v ...interface{}) {
	l.t.Log(v...)
}

func (l *TestLogger) Printf(format string, v ...interface{}) {
	l.t.Logf(format, v...)
}

func shouldInstallPDSHelmChart(versionConstraints string) bool {
	return versionConstraints != "0"
}
