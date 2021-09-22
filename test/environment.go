package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	envControlPlaneAPI = "CONTROL_PLANE_API"
	envTargetAPI       = "TARGET_API"
	envTargetToken     = "TARGET_TOKEN"
)

func mustGetEnvVariable(t *testing.T, key string) string {
	t.Helper()

	value := os.Getenv(key)
	require.NotEmptyf(t, value, "Env variable %q is empty.", key)
	return value
}
