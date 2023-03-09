package test

import (
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/api"
)

const (
	envControlPlaneAddress      = "CONTROL_PLANE_ADDRESS"
	envPDSAccountName           = "PDS_ACCOUNT_NAME"
	envPDSTenantName            = "PDS_TENANT_NAME"
	envPDSProjectName           = "PDS_PROJECT_NAME"
	envPDSDeploymentTargetName  = "PDS_DEPTARGET_NAME"
	envPDSNamespaceName         = "PDS_NAMESPACE_NAME"
	envPDSServiceAccountName    = "PDS_SERVICE_ACCOUNT_NAME"
	envTargetKubeconfig         = "TARGET_CLUSTER_KUBECONFIG"
	envSecretTokenIssuerURL     = "SECRET_TOKEN_ISSUER_URL"
	envSecretIssuerClientID     = "SECRET_ISSUER_CLIENT_ID"
	envSecretIssuerClientSecret = "SECRET_ISSUER_CLIENT_SECRET"
	envSecretPDSUsername        = "SECRET_PDS_USERNAME"
	envSecretPDSPassword        = "SECRET_PDS_PASSWORD"
	envBackupTargetBucket       = "PDS_BACKUPTARGET_BUCKET"
	envBackupTargetRegion       = "PDS_BACKUPTARGET_REGION"
	envS3CredentialsAccessKey   = "PDS_S3CREDENTIALS_ACCESSKEY"
	envS3CredentialsEndpoint    = "PDS_S3CREDENTIALS_ENDPOINT"
	envS3CredentialsSecretKey   = "PDS_S3CREDENTIALS_SECRETKEY"
	envPDSToken                 = "SECRET_PDS_TOKEN"
	envPDSHelmChartVersion      = "PDS_HELM_CHART_VERSION"
)

const (
	defaultPDSServiceAccountName = "Default-AgentWriter"
	defaultPDSTenantName         = "Default"
	defaultPDSProjectName        = "Default"
	defaultPDSNamespaceName      = "dev"
	defaultPDSAccountName        = "PDS Integration Test"
	defaultS3Endpoint            = "s3.amazonaws.com"
)

var (
	// runTimestamp is the current timestamp to identify resources created within a single test run.
	runTimestamp = time.Now().Format("2006-01-02 15:04:05")

	defaultPDSDeploymentTargetName = "PDS Integration Test Cluster " + runTimestamp
)

type backupCredentials struct {
	s3 s3Credentials
}

type s3Credentials struct {
	accessKey string
	endpoint  string
	secretKey string
}

type backupTargetConfig struct {
	bucket      string
	region      string
	credentials backupCredentials
}

type controlPlaneEnvironment struct {
	ControlPlaneAPI  string
	AccountName      string
	TenantName       string
	ProjectName      string
	LoginCredentials api.LoginCredentials
	PrometheusAPI    string
}

type environment struct {
	controlPlane            controlPlaneEnvironment
	targetKubeconfig        string
	pdsNamespaceName        string
	pdsDeploymentTargetName string
	pdsServiceAccountName   string
	pdsToken                string
	backupTarget            backupTargetConfig
	pdsHelmChartVersion     string
}

func MustHaveControlPlaneEnvVariables(t *testing.T) controlPlaneEnvironment {
	credentials := api.LoginCredentials{
		BearerToken: os.Getenv(envPDSToken),
	}
	if credentials.BearerToken == "" {
		credentials.TokenIssuerURL = mustGetEnvVariable(t, envSecretTokenIssuerURL)
		credentials.IssuerClientID = mustGetEnvVariable(t, envSecretIssuerClientID)
		credentials.IssuerClientSecret = mustGetEnvVariable(t, envSecretIssuerClientSecret)
		credentials.Username = mustGetEnvVariable(t, envSecretPDSUsername)
		credentials.Password = mustGetEnvVariable(t, envSecretPDSPassword)
	}

	controlPlaneAddress := mustCleanAddress(t, mustGetEnvVariable(t, envControlPlaneAddress))
	return controlPlaneEnvironment{
		ControlPlaneAPI:  fmt.Sprintf("https://%s/api", controlPlaneAddress),
		AccountName:      getEnvVariableWithDefault(envPDSAccountName, defaultPDSAccountName),
		TenantName:       getEnvVariableWithDefault(envPDSTenantName, defaultPDSTenantName),
		ProjectName:      getEnvVariableWithDefault(envPDSProjectName, defaultPDSProjectName),
		LoginCredentials: credentials,
		PrometheusAPI:    fmt.Sprintf("https://%s/prometheus", controlPlaneAddress),
	}
}

func mustHaveEnvVariables(t *testing.T) environment {
	t.Helper()

	pdsToken := os.Getenv(envPDSToken)
	var authConf api.LoginCredentials
	if pdsToken == "" {
		authConf = api.LoginCredentials{
			TokenIssuerURL:     mustGetEnvVariable(t, envSecretTokenIssuerURL),
			IssuerClientID:     mustGetEnvVariable(t, envSecretIssuerClientID),
			IssuerClientSecret: mustGetEnvVariable(t, envSecretIssuerClientSecret),
			Username:           mustGetEnvVariable(t, envSecretPDSUsername),
			Password:           mustGetEnvVariable(t, envSecretPDSPassword),
		}
	}

	controlPlaneAddress := mustGetEnvVariable(t, envControlPlaneAddress)
	controlPlaneAddress = mustCleanAddress(t, controlPlaneAddress)

	return environment{
		controlPlane: controlPlaneEnvironment{
			ControlPlaneAPI:  fmt.Sprintf("https://%s/api", controlPlaneAddress),
			AccountName:      getEnvVariableWithDefault(envPDSAccountName, defaultPDSAccountName),
			TenantName:       getEnvVariableWithDefault(envPDSTenantName, defaultPDSTenantName),
			ProjectName:      getEnvVariableWithDefault(envPDSProjectName, defaultPDSProjectName),
			LoginCredentials: authConf,
			PrometheusAPI:    fmt.Sprintf("https://%s/prometheus", controlPlaneAddress),
		},
		targetKubeconfig:        mustGetEnvVariable(t, envTargetKubeconfig),
		pdsNamespaceName:        getEnvVariableWithDefault(envPDSNamespaceName, defaultPDSNamespaceName),
		pdsDeploymentTargetName: getEnvVariableWithDefault(envPDSDeploymentTargetName, defaultPDSDeploymentTargetName),
		pdsServiceAccountName:   getEnvVariableWithDefault(envPDSServiceAccountName, defaultPDSServiceAccountName),
		pdsToken:                pdsToken,
		backupTarget: backupTargetConfig{
			bucket: mustGetEnvVariable(t, envBackupTargetBucket),
			region: mustGetEnvVariable(t, envBackupTargetRegion),
			credentials: backupCredentials{
				s3: s3Credentials{
					accessKey: mustGetEnvVariable(t, envS3CredentialsAccessKey),
					endpoint:  getEnvVariableWithDefault(envS3CredentialsEndpoint, defaultS3Endpoint),
					secretKey: mustGetEnvVariable(t, envS3CredentialsSecretKey),
				},
			},
		},
		pdsHelmChartVersion: getEnvVariableWithDefault(envPDSHelmChartVersion, ""),
	}
}

func mustGetEnvVariable(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	require.NotEmptyf(t, value, "Env variable %q is empty.", key)
	return value
}

func mustCleanAddress(t *testing.T, address string) string {
	url, err := url.Parse(address)
	require.NoError(t, err, "failed to parse address")
	if url.Hostname() == "" {
		return url.String()
	}
	return url.Hostname()
}

func getEnvVariableWithDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
