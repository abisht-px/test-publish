package test

import (
	"context"
	"fmt"
	"io"
	"testing"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
	batchv1 "k8s.io/api/batch/v1"
)

const (
	pdsDeploymentTargetHealthState = "healthy"
	pdsDeploymentHealthState       = "Healthy"
)

func isDeploymentTargetHealthy(ctx context.Context, apiClient *pds.APIClient, deploymentTargetID string) bool {
	target, _, err := apiClient.DeploymentTargetsApi.ApiDeploymentTargetsIdGet(ctx, deploymentTargetID).Execute()
	if err != nil {
		return false
	}
	if target.GetStatus() != pdsDeploymentTargetHealthState {
		return false
	}
	return true
}

func getDeploymentTargetIDByName(ctx context.Context, apiClient *pds.APIClient, tenantID, deploymentTargetName string) (string, error) {
	targets, _, err := apiClient.DeploymentTargetsApi.ApiTenantsIdDeploymentTargetsGet(ctx, tenantID).Execute()

	if err != nil {
		return "", err
	}

	for _, target := range targets.GetData() {
		if target.GetName() == deploymentTargetName {
			return target.GetId(), nil
		}
	}
	return "", fmt.Errorf("deployment target %s not found", deploymentTargetName)
}

func getNamespaceIDByName(ctx context.Context, apiClient *pds.APIClient, deploymentTargetID, namespaceName string) (string, error) {
	namespaces, _, err := apiClient.NamespacesApi.ApiDeploymentTargetsIdNamespacesGet(ctx, deploymentTargetID).Execute()

	if err != nil {
		return "", err
	}

	for _, namespace := range namespaces.GetData() {
		if namespace.GetName() == namespaceName {
			return namespace.GetId(), nil
		}
	}
	return "", fmt.Errorf("namespace %s not found", namespaceName)
}

func isDeploymentHealthy(ctx context.Context, apiClient *pds.APIClient, deploymentID string) bool {
	deployment, _, err := apiClient.DeploymentsApi.ApiDeploymentsIdStatusGet(ctx, deploymentID).Execute()
	if err != nil {
		return false
	}
	if deployment.GetHealth() != pdsDeploymentHealthState {
		return false
	}
	return true
}

func getAllImageVersions(ctx context.Context, apiClient *pds.APIClient) ([]PDSImageReferenceSpec, error) {
	var records []PDSImageReferenceSpec

	dataServices, _, err := apiClient.DataServicesApi.ApiDataServicesGet(ctx).Execute()
	if err != nil {
		return nil, err
	}

	for _, dataService := range dataServices.GetData() {
		versions, _, err := apiClient.VersionsApi.ApiDataServicesIdVersionsGet(ctx, dataService.GetId()).Execute()
		if err != nil {
			return nil, err
		}

		for _, version := range versions.GetData() {
			images, _, err := apiClient.ImagesApi.ApiVersionsIdImagesGet(ctx, version.GetId()).Execute()
			if err != nil {
				return nil, err
			}

			for _, image := range images.GetData() {
				record := PDSImageReferenceSpec{
					ServiceName:       dataService.GetName(),
					DataServiceID:     dataService.GetId(),
					VersionID:         version.GetId(),
					ImageVersionBuild: image.GetBuild(),
					ImageVersionTag:   image.GetTag(),
					ImageID:           image.GetId(),
				}
				records = append(records, record)
			}
		}
	}

	return records, nil
}

func findImageVersionForRecord(deployment *ShortDeploymentSpec, images []PDSImageReferenceSpec) *PDSImageReferenceSpec {
	for _, image := range images {
		found := image.ServiceName == deployment.ServiceName
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

func createPDSDeployment(ctx context.Context, apiClient *pds.APIClient, deployment *ShortDeploymentSpec, image *PDSImageReferenceSpec, teanntID, deploymentTargetID, projectID, namespaceID string) (string, error) {
	resources, _, err := apiClient.ResourceSettingsTemplatesApi.ApiTenantsIdResourceSettingsTemplatesGet(ctx, teanntID).Name(deployment.ResourceSettingsTemplateName).Execute()
	if err != nil {
		return "", err
	}
	var resource *pds.ModelsResourceSettingsTemplate
	for _, r := range resources.GetData() {
		if r.GetDataServiceId() == image.DataServiceID {
			resource = &r
			break
		}
	}
	if resource == nil {
		return "", fmt.Errorf("resource settings template %s not found", deployment.ResourceSettingsTemplateName)
	}

	storages, _, err := apiClient.StorageOptionsTemplatesApi.ApiTenantsIdStorageOptionsTemplatesGet(ctx, teanntID).Name(deployment.StorageOptionName).Execute()
	if err != nil {
		return "", err
	}
	if len(storages.GetData()) == 0 {
		return "", fmt.Errorf("storage option template %s not found", deployment.StorageOptionName)
	}
	if len(storages.GetData()) != 1 {
		return "", fmt.Errorf("more than one storage option template found")
	}
	storage := storages.GetData()[0]

	appConfigurations, _, err := apiClient.ApplicationConfigurationTemplatesApi.ApiTenantsIdApplicationConfigurationTemplatesGet(ctx, teanntID).Name(deployment.AppConfigTemplateName).Execute()
	if err != nil {
		return "", err
	}
	var appConfig *pds.ModelsApplicationConfigurationTemplate
	for _, c := range appConfigurations.GetData() {
		if c.GetDataServiceId() == image.DataServiceID {
			appConfig = &c
			break
		}
	}
	if appConfig == nil {
		return "", fmt.Errorf("application configuration template %s not found", deployment.AppConfigTemplateName)
	}

	var backupPolicy *pds.ModelsBackupPolicy
	if len(deployment.BackupPolicyname) > 0 {
		backupPolicies, _, err := apiClient.BackupPoliciesApi.ApiTenantsIdBackupPoliciesGet(ctx, teanntID).Name(deployment.BackupPolicyname).Execute()
		if err != nil {
			return "", err
		}
		if len(backupPolicies.GetData()) == 0 {
			return "", fmt.Errorf("backup policy %s not found", deployment.BackupPolicyname)
		}
		if len(backupPolicies.GetData()) != 1 {
			return "", fmt.Errorf("more than one backup policy found")
		}
		backupPolicy = &backupPolicies.GetData()[0]
	}

	dns, _, err := apiClient.TenantsApi.ApiTenantsIdDnsDetailsGet(ctx, teanntID).Execute()
	if err != nil {
		return "", err
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

func getDeploymentPodName(deploymentName string) string {
	return fmt.Sprintf("%s-0", deploymentName)
}

type TestLogger struct {
	t *testing.T
}

func newTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{t}
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
