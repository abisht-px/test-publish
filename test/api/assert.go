package api

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func RequireNoError(t *testing.T, resp *http.Response, err error) {
	t.Helper()

	if err != nil {
		rawbody, parseErr := io.ReadAll(resp.Body)
		require.NoError(t, parseErr, "Error calling PDS API: failed to read response body")
		require.NoErrorf(t, err, "Error calling PDS API: %s", rawbody)
	}
}

func RequireNoErrorf(t *testing.T, resp *http.Response, err error, msg string, args ...any) {
	t.Helper()

	if err != nil {
		rawbody, parseErr := io.ReadAll(resp.Body)
		require.NoError(t, parseErr, "Error calling PDS API: failed to read response body")
		details := fmt.Sprintf(msg, args...)
		require.NoErrorf(t, err, "Error calling PDS API: %s: %s", details, rawbody)
	}
}
