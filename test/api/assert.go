package api

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func NoError(t *testing.T, resp *http.Response, err error) bool {
	t.Helper()
	if err == nil {
		return true
	}
	rawbody, parseErr := io.ReadAll(resp.Body)
	assert.NoError(t, parseErr, "Error calling PDS API: failed to read response body")
	assert.NoErrorf(t, err, "Error calling PDS API: %s", rawbody)
	return false
}

func NoErrorf(t *testing.T, resp *http.Response, err error, msg string, args ...any) bool {
	t.Helper()
	if err == nil {
		return true
	}
	rawbody, parseErr := io.ReadAll(resp.Body)
	assert.NoError(t, parseErr, "Error calling PDS API: failed to read response body")
	details := fmt.Sprintf(msg, args...)
	assert.NoErrorf(t, err, "Error calling PDS API: %s: %s", details, rawbody)
	return false
}

func RequireNoError(t *testing.T, resp *http.Response, err error) {
	t.Helper()
	if NoError(t, resp, err) {
		return
	}
	t.FailNow()
}

func RequireNoErrorf(t *testing.T, resp *http.Response, err error, msg string, args ...any) {
	t.Helper()
	if NoErrorf(t, resp, err, msg, args...) {
		return
	}
	t.FailNow()
}
