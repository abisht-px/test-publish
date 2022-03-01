package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	envControlPlaneAPI        = "CONTROL_PLANE_API"
	envBearerToken            = "CONTROL_PLANE_BEARER_TOKEN"
	envControlPlaneKubeconfig = "CONTROL_PLANE_KUBECONFIG"
	envTargetKubeconfig       = "TARGET_KUBECONFIG"
)

type environment struct {
	controlPlaneAPI, bearerToken, controlPlaneKubeconfig, targetKubeconfig string
}

func mustHaveEnvVariables(t *testing.T) environment {
	t.Helper()
	return environment{
		controlPlaneAPI:        mustGetEnvVariable(t, envControlPlaneAPI),
		bearerToken:            mustGetEnvVariable(t, envBearerToken),
		controlPlaneKubeconfig: mustGetEnvVariable(t, envControlPlaneKubeconfig),
		targetKubeconfig:       mustGetEnvVariable(t, envTargetKubeconfig),
	}
}

func mustGetEnvVariable(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	require.NotEmptyf(t, value, "Env variable %q is empty.", key)
	return value
}
