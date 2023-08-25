package framework

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/lithammer/shortuuid/v3"
	"github.com/pkg/errors"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/internal/random"
	"github.com/portworx/pds-integration-test/internal/tests"
)

func NewLoginCredentialsFromFlags() api.LoginCredentials {
	return api.LoginCredentials{
		TokenIssuerURL:     IssuerTokenURL,
		IssuerClientID:     IssuerClientID,
		IssuerClientSecret: IssuerClientSecret,
		Username:           PDSUsername,
		Password:           PDSPassword,
		BearerToken:        PDSAPIToken,
	}
}

func NewControlPlane(
	t tests.T,
	apiClient *api.PDSClient,
	opts ...controlplane.InitializeOption,
) *controlplane.ControlPlane {
	cp := controlplane.New(apiClient)

	for _, o := range opts {
		o(context.Background(), t, cp)
	}

	return cp
}

func NewPDSChartConfigFromFlags(
	tenantID string,
	serviceAccountToken string,
	controlPlaneAPI string,
) targetcluster.PDSChartConfig {
	return targetcluster.PDSChartConfig{
		Version:               PDSHelmChartVersion,
		TenantID:              tenantID,
		Token:                 serviceAccountToken,
		ControlPlaneAPI:       controlPlaneAPI,
		DeploymentTargetName:  DeploymentTargetName,
		DataServiceTLSEnabled: DataServiceTLSEnabled,
	}
}

func NewCertManagerChartConfigFromFlags() targetcluster.CertManagerChartConfig {
	return targetcluster.CertManagerChartConfig{
		Version: CertManagerChartVersion,
	}
}

func NewTargetClusterFromFlags(tenantID, serviceAccountToken string) (*targetcluster.TargetCluster, error) {
	pdsChartCfg := NewPDSChartConfigFromFlags(tenantID, serviceAccountToken, PDSControlPlaneAPI)
	certManagerChartCfg := NewCertManagerChartConfigFromFlags()

	tc, err := targetcluster.NewTargetCluster(
		context.Background(),
		TargetClusterKubeconfig,
		pdsChartCfg,
		certManagerChartCfg,
	)
	if err != nil {
		return nil, errors.Wrap(err, "initialize target cluster")
	}

	return tc, nil
}

func NewBackupCredentialFromFlags() controlplane.BackupCredentials {
	return controlplane.BackupCredentials{
		S3: controlplane.S3Credentials{
			AccessKey: AWSAccessKey,
			SecretKey: AWSSecretKey,
			Endpoint:  AWSS3Endpoint,
		},
	}
}

func NewBackupTargetConfigFromFlags() BackupTargetConfig {
	return BackupTargetConfig{
		Bucket:      AWSS3BucketName,
		Region:      AWSRegion,
		Credentials: NewBackupCredentialFromFlags(),
	}
}

func ShouldRegister() bool {
	return PDSHelmChartVersion != "0"
}

func InitializePDSHelmChartVersion(t *testing.T, apiClient *api.PDSClient) {
	if PDSHelmChartVersion != "" {
		return
	}

	metadata, resp, err := apiClient.MetadataApi.ApiMetadataGet(context.Background()).Execute()
	api.RequireNoError(t, resp, err)

	PDSHelmChartVersion = strings.TrimPrefix(metadata.GetHelmChartVersion(), "v")
}

func NewShortUUID() string {
	return shortuuid.New()
}

func NewRandomName(prefix string) string {
	return fmt.Sprintf("ft-%s-%s", prefix, random.AlphaNumericString(random.NameSuffixLength))
}

func ErrorContainsAnyOfMsg(err error, msgs ...string) bool {
	for _, msg := range msgs {
		if err != nil && strings.Contains(err.Error(), msg) {
			return true
		}
	}

	return false
}
