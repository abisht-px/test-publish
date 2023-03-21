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
	pdsDeploymentHealthState = "Healthy"
)

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
