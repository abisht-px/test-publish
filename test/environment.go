package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	envControlPlaneAPI        = "CONTROL_PLANE_API"
	envControlPlaneKubeconfig = "CONTROL_PLANE_KUBECONFIG"
	envTargetAPI              = "TARGET_API"
	envTargetToken            = "TARGET_TOKEN"
	envTargetKubeconfig       = "TARGET_KUBECONFIG"
)

type environment struct {
	controlPlaneAPI, controlPlaneKubeconfig, targetAPI, targetToken, targetKubeconfig string
}

func mustHaveEnvVariables(t *testing.T) environment {
	t.Helper()
	return environment{
		controlPlaneAPI:        mustGetEnvVariable(t, envControlPlaneAPI),
		controlPlaneKubeconfig: mustGetEnvVariable(t, envControlPlaneKubeconfig),
		targetAPI:              mustGetEnvVariable(t, envTargetAPI),
		targetToken:            mustGetEnvVariable(t, envTargetToken),
		targetKubeconfig:       mustGetEnvVariable(t, envTargetKubeconfig),
	}
}

func mustGetEnvVariable(t *testing.T, key string) string {
	t.Helper()

	value := os.Getenv(key)
	require.NotEmptyf(t, value, "Env variable %q is empty.", key)
	return value
}
