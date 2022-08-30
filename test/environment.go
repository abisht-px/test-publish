package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	envControlPlaneAPI           = "CONTROL_PLANE_API"
	envPDSAccountName            = "PDS_ACCOUNT_NAME"
	envPDSTenantName             = "PDS_TENANT_NAME"
	envPDSProjectName            = "PDS_PROJECT_NAME"
	envPDSDeploymentTargetName   = "PDS_DEPTARGET_NAME"
	envPDSNamespaceName          = "PDS_NAMESPACE_NAME"
	envPDSServiceAccountName     = "PDS_SERVICE_ACCOUNT_NAME"
	envTargetKubeconfig          = "TARGET_CLUSTER_KUBECONFIG"
	envSecretTokenIssuerURL      = "SECRET_TOKEN_ISSUER_URL"
	envSecretIssuerClientID      = "SECRET_ISSUER_CLIENT_ID"
	envSecretIssuerClientSecret  = "SECRET_ISSUER_CLIENT_SECRET"
	envSecretPDSUsername         = "SECRET_PDS_USERNAME"
	envSecretPDSPassword         = "SECRET_PDS_PASSWORD"
	envShortDeploymentSpecPrefix = "PDS_DEPLOYMENT_SPEC"
)

const (
	defaultPDSServiceAccountName   = "Default-AgentWriter"
	defaultPDSTenantName           = "Default"
	defaultPDSProjectName          = "Default"
	defaultPDSNamespaceName        = "dev"
	defaultPDSDeploymentTargetName = "PDS Integration Test Cluster"
	defaultPDSAccountName          = "PDS Integration Test"
)

type secrets struct {
	tokenIssuerURL     string
	issuerClientID     string
	issuerClientSecret string
	pdsUsername        string
	pdsPassword        string
}

type environment struct {
	controlPlaneAPI         string
	targetKubeconfig        string
	pdsAccountName          string
	pdsTenantName           string
	pdsProjectName          string
	pdsNamespaceName        string
	pdsDeploymentTargetName string
	pdsServiceAccountName   string
	secrets                 secrets
}

func mustHaveEnvVariables(t *testing.T) environment {
	t.Helper()
	return environment{
		controlPlaneAPI:         mustGetEnvVariable(t, envControlPlaneAPI),
		targetKubeconfig:        mustGetEnvVariable(t, envTargetKubeconfig),
		pdsAccountName:          getEnvVariableWithDefault(envPDSAccountName, defaultPDSAccountName),
		pdsTenantName:           getEnvVariableWithDefault(envPDSTenantName, defaultPDSTenantName),
		pdsProjectName:          getEnvVariableWithDefault(envPDSProjectName, defaultPDSProjectName),
		pdsNamespaceName:        getEnvVariableWithDefault(envPDSNamespaceName, defaultPDSNamespaceName),
		pdsDeploymentTargetName: getEnvVariableWithDefault(envPDSDeploymentTargetName, defaultPDSDeploymentTargetName),
		pdsServiceAccountName:   getEnvVariableWithDefault(envPDSServiceAccountName, defaultPDSServiceAccountName),
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

func mustGetEnvList(t *testing.T, key string) []string {
	t.Helper()
	list := make([]string, 0)
	index := 0
	for {
		envVarName := fmt.Sprintf("%s%d", key, index)
		value, ok := os.LookupEnv(envVarName)
		if !ok {
			break
		}
		list = append(list, value)
		index++
	}

	require.NotEmptyf(t, list, "Env variable %q is empty.", key)
	return list
}

func getEnvVariableWithDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
