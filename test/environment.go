package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	envControlPlaneAPI          = "CONTROL_PLANE_API"
	envTargetKubeconfig         = "TARGET_CLUSTER_KUBECONFIG"
	envSecretTokenIssuerURL     = "SECRET_TOKEN_ISSUER_URL"
	envSecretIssuerClientID     = "SECRET_ISSUER_CLIENT_ID"
	envSecretIssuerClientSecret = "SECRET_ISSUER_CLIENT_SECRET"
	envSecretPDSUsername        = "SECRET_PDS_USERNAME"
	envSecretPDSPassword        = "SECRET_PDS_PASSWORD"
)

type secrets struct {
	tokenIssuerURL     string
	issuerClientID     string
	issuerClientSecret string
	pdsUsername        string
	pdsPassword        string
}

type environment struct {
	controlPlaneAPI  string
	targetKubeconfig string
	secrets          secrets
}

func mustHaveEnvVariables(t *testing.T) environment {
	t.Helper()
	return environment{
		controlPlaneAPI:  mustGetEnvVariable(t, envControlPlaneAPI),
		targetKubeconfig: mustGetEnvVariable(t, envTargetKubeconfig),
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
