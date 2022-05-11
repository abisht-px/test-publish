package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	envControlPlaneAPI  = "CONTROL_PLANE_API"
	envTargetKubeconfig = "TARGET_CLUSTER_KUBECONFIG"
)

type environment struct {
	controlPlaneAPI  string
	targetKubeconfig string
}

func mustHaveEnvVariables(t *testing.T) environment {
	t.Helper()
	return environment{
		controlPlaneAPI:  mustGetEnvVariable(t, envControlPlaneAPI),
		targetKubeconfig: mustGetEnvVariable(t, envTargetKubeconfig),
	}
}

func mustGetEnvVariable(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	require.NotEmptyf(t, value, "Env variable %q is empty.", key)
	return value
}
