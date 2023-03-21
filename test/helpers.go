package test

import (
	"context"
	"fmt"
	"testing"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
	batchv1 "k8s.io/api/batch/v1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/random"
)

const (
	pdsDeploymentTargetHealthState = "healthy"
	pdsDeploymentHealthState       = "Healthy"
)

func checkDeploymentTargetHealth(ctx context.Context, apiClient *api.PDSClient, deploymentTargetID string) error {
	target, resp, err := apiClient.DeploymentTargetsApi.ApiDeploymentTargetsIdGet(ctx, deploymentTargetID).Execute()
	if err != nil {
		return api.ExtractErrorDetails(resp, err)
	}
	if target.GetStatus() != pdsDeploymentTargetHealthState {
		return fmt.Errorf("deployment target not healthy: got %q, want %q", target.GetStatus(), pdsDeploymentTargetHealthState)
	}
	return nil
}

func getDeploymentTargetIDByName(ctx context.Context, apiClient *api.PDSClient, tenantID, deploymentTargetName string) (string, error) {
	targets, resp, err := apiClient.DeploymentTargetsApi.ApiTenantsIdDeploymentTargetsGet(ctx, tenantID).Execute()
	if err = api.ExtractErrorDetails(resp, err); err != nil {
		return "", fmt.Errorf("getting deployment targets for tenant %s: %w", tenantID, err)
	}
	for _, target := range targets.GetData() {
		if target.GetName() == deploymentTargetName {
			return target.GetId(), nil
		}
	}
	return "", fmt.Errorf("deployment target %s not found", deploymentTargetName)
}

func getNamespaceByName(ctx context.Context, apiClient *api.PDSClient, deploymentTargetID, name string) (*pds.ModelsNamespace, error) {
	namespaces, resp, err := apiClient.NamespacesApi.ApiDeploymentTargetsIdNamespacesGet(ctx, deploymentTargetID).Execute()
	if err = api.ExtractErrorDetails(resp, err); err != nil {
		return nil, fmt.Errorf("getting namespace %s: %w", name, err)
	}
	for _, namespace := range namespaces.GetData() {
		if namespace.GetName() == name {
			return &namespace, nil
		}
	}
	return nil, nil
}

func getAllImageVersions(ctx context.Context, apiClient *api.PDSClient) ([]api.PDSImageReferenceSpec, error) {
	var records []api.PDSImageReferenceSpec

	dataServices, resp, err := apiClient.DataServicesApi.ApiDataServicesGet(ctx).Execute()
	if err = api.ExtractErrorDetails(resp, err); err != nil {
		return nil, fmt.Errorf("fetching all data services: %w", err)
	}

	dataServicesByID := make(map[string]pds.ModelsDataService)
	for i := range dataServices.GetData() {
		dataService := dataServices.GetData()[i]
		dataServicesByID[dataService.GetId()] = dataService
	}

	images, resp, err := apiClient.ImagesApi.ApiImagesGet(ctx).Latest(true).SortBy("-created_at").Limit("1000").Execute()
	if err = api.ExtractErrorDetails(resp, err); err != nil {
		return nil, fmt.Errorf("fetching all images: %w", err)
	}

	for _, image := range images.GetData() {
		dataService := dataServicesByID[image.GetDataServiceId()]
		record := api.PDSImageReferenceSpec{
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

func createPDSDeployment(ctx context.Context, apiClient *api.PDSClient, deployment *api.ShortDeploymentSpec, image *api.PDSImageReferenceSpec, tenantID, deploymentTargetID, projectID, namespaceID string) (string, error) {
	resource, err := apiClient.GetResourceSettingsTemplateByName(ctx, tenantID, deployment.ResourceSettingsTemplateName, image.DataServiceID)
	if err != nil {
		return "", fmt.Errorf("getting resource settings template %s for tenant %s: %w", deployment.ResourceSettingsTemplateName, tenantID, err)
	}

	appConfig, err := apiClient.GetAppConfigTemplateByName(ctx, tenantID, deployment.AppConfigTemplateName, image.DataServiceID)
	if err != nil {
		return "", fmt.Errorf("getting application configuration template %s for tenant %s: %w", deployment.AppConfigTemplateName, tenantID, err)
	}

	storages, resp, err := apiClient.StorageOptionsTemplatesApi.ApiTenantsIdStorageOptionsTemplatesGet(ctx, tenantID).Name(deployment.StorageOptionName).Execute()
	if err = api.ExtractErrorDetails(resp, err); err != nil {
		return "", fmt.Errorf("getting storage option template %s for tenant %s: %w", deployment.StorageOptionName, tenantID, err)
	}

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
		if err = api.ExtractErrorDetails(resp, err); err != nil {
			return "", fmt.Errorf("getting backup policies for tenant %s: %w", tenantID, err)
		}
		if len(backupPolicies.GetData()) == 0 {
			return "", fmt.Errorf("backup policy %s not found", deployment.BackupPolicyname)
		}
		if len(backupPolicies.GetData()) != 1 {
			return "", fmt.Errorf("more than one backup policy found")
		}
		backupPolicy = &backupPolicies.GetData()[0]
	}

	dns, resp, err := apiClient.TenantsApi.ApiTenantsIdDnsDetailsGet(ctx, tenantID).Execute()
	if err = api.ExtractErrorDetails(resp, err); err != nil {
		return "", fmt.Errorf("getting DNS details for tenant %s: %w", tenantID, err)
	}

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
	if err = api.ExtractErrorDetails(httpRes, err); err != nil {
		return "", fmt.Errorf("deploying %s under project %s: %w", *pdsDeployment.Name, projectID, err)
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

func generateRandomName(prefix string) string {
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	return fmt.Sprintf("%s-integration-test-s3-%s", prefix, nameSuffix)
}
