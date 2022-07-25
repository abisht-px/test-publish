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
	envPDSServiceAccountName    = "PDS_SERVICE_ACCOUNT_NAME"
	envTargetKubeconfig         = "TARGET_CLUSTER_KUBECONFIG"
	envSecretTokenIssuerURL     = "SECRET_TOKEN_ISSUER_URL"
	envSecretIssuerClientID     = "SECRET_ISSUER_CLIENT_ID"
	envSecretIssuerClientSecret = "SECRET_ISSUER_CLIENT_SECRET"
	envSecretPDSUsername        = "SECRET_PDS_USERNAME"
	envSecretPDSPassword        = "SECRET_PDS_PASSWORD"
)

const (
	defaultPDSServiceAccountName = "Default-AgentWriter"
	defaultPDSTenantName         = "Default"
	defaultPDSAccountName        = "PDS Integration Test"
)

type secrets struct {
	tokenIssuerURL     string
	issuerClientID     string
	issuerClientSecret string
	pdsUsername        string
	pdsPassword        string
}

type environment struct {
	controlPlaneAPI       string
	targetKubeconfig      string
	pdsAccountName        string
	pdsTenantName         string
	pdsServiceAccountName string
	secrets               secrets
}

func mustHaveEnvVariables(t *testing.T) environment {
	t.Helper()
	return environment{
		controlPlaneAPI:       mustGetEnvVariable(t, envControlPlaneAPI),
		targetKubeconfig:      mustGetEnvVariable(t, envTargetKubeconfig),
		pdsAccountName:        getEnvVariableWithDefault(envPDSAccountName, defaultPDSAccountName),
		pdsTenantName:         getEnvVariableWithDefault(envPDSTenantName, defaultPDSTenantName),
		pdsServiceAccountName: getEnvVariableWithDefault(envPDSServiceAccountName, defaultPDSServiceAccountName),
		secrets: secrets{
			tokenIssuerURL:     mustGetEnvVariable(t, envSecretTokenIssuerURL),
			issuerClientID:     mustGetEnvVariable(t, envSecretIssuerClientID),
			issuerClientSecret: mustGetEnvVariable(t, envSecretIssuerClientSecret),
			pdsUsername:        mustGetEnvVariable(t, envSecretPDSUsername),
			pdsPassword:        mustGetEnvVariable(t, envSecretPDSPassword),
		},
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
