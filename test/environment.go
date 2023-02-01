package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	envControlPlaneAPI          = "CONTROL_PLANE_API"
	envPDSAccountName           = "PDS_ACCOUNT_NAME"
	envPDSTenantName            = "PDS_TENANT_NAME"
	envPDSProjectName           = "PDS_PROJECT_NAME"
	envPDSDeploymentTargetName  = "PDS_DEPTARGET_NAME"
	envPDSNamespaceName         = "PDS_NAMESPACE_NAME"
	envPXNamespaceName          = "PX_NAMESPACE_NAME"
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
	defaultPDSServiceAccountName   = "Default-AgentWriter"
	defaultPDSTenantName           = "Default"
	defaultPDSProjectName          = "Default"
	defaultPDSNamespaceName        = "dev"
	defaultPXNamespaceName         = "kube-system"
	defaultPDSDeploymentTargetName = "PDS Integration Test Cluster"
	defaultPDSAccountName          = "PDS Integration Test"
	defaultS3Endpoint              = "s3.amazonaws.com"
)

type secrets struct {
	tokenIssuerURL     string
	issuerClientID     string
	issuerClientSecret string
	pdsUsername        string
	pdsPassword        string
}

type credentials struct {
	s3 s3Credentials
}

type s3Credentials struct {
	accessKey string
	endpoint  string
	secretKey string
}

type backupTarget struct {
	bucket string
	region string
	credentials
}

type environment struct {
	controlPlaneAPI         string
	targetKubeconfig        string
	pdsAccountName          string
	pdsTenantName           string
	pdsProjectName          string
	pdsNamespaceName        string
	pxNamespaceName         string
	pdsDeploymentTargetName string
	pdsServiceAccountName   string
	pdsToken                string
	secrets                 secrets
	backupTarget            backupTarget
	pdsHelmChartVersion     string
}

func mustHaveEnvVariables(t *testing.T) environment {
	t.Helper()

	pdsToken := os.Getenv(envPDSToken)
	var authConf secrets
	if pdsToken == "" {
		authConf = secrets{
			tokenIssuerURL:     mustGetEnvVariable(t, envSecretTokenIssuerURL),
			issuerClientID:     mustGetEnvVariable(t, envSecretIssuerClientID),
			issuerClientSecret: mustGetEnvVariable(t, envSecretIssuerClientSecret),
			pdsUsername:        mustGetEnvVariable(t, envSecretPDSUsername),
			pdsPassword:        mustGetEnvVariable(t, envSecretPDSPassword),
		}
	}

	return environment{
		controlPlaneAPI:         mustGetEnvVariable(t, envControlPlaneAPI),
		targetKubeconfig:        mustGetEnvVariable(t, envTargetKubeconfig),
		pdsAccountName:          getEnvVariableWithDefault(envPDSAccountName, defaultPDSAccountName),
		pdsTenantName:           getEnvVariableWithDefault(envPDSTenantName, defaultPDSTenantName),
		pdsProjectName:          getEnvVariableWithDefault(envPDSProjectName, defaultPDSProjectName),
		pdsNamespaceName:        getEnvVariableWithDefault(envPDSNamespaceName, defaultPDSNamespaceName),
		pxNamespaceName:         getEnvVariableWithDefault(envPXNamespaceName, defaultPXNamespaceName),
		pdsDeploymentTargetName: getEnvVariableWithDefault(envPDSDeploymentTargetName, defaultPDSDeploymentTargetName),
		pdsServiceAccountName:   getEnvVariableWithDefault(envPDSServiceAccountName, defaultPDSServiceAccountName),
		pdsToken:                pdsToken,
		secrets:                 authConf,
		backupTarget: backupTarget{
			bucket: mustGetEnvVariable(t, envBackupTargetBucket),
			region: mustGetEnvVariable(t, envBackupTargetRegion),
			credentials: credentials{
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

func getEnvVariableWithDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
