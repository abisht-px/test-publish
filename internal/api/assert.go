package api

import (
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"

	"github.com/portworx/pds-integration-test/internal/tests"
)

func NoError(t tests.T, resp *http.Response, err error) bool {
	t.Helper()
	apierr := ExtractErrorDetails(resp, err)
	return assert.NoError(t, apierr, "Error calling PDS API.")
}

func NoErrorf(t tests.T, resp *http.Response, err error, msg string, args ...any) bool {
	t.Helper()
	apierr := ExtractErrorDetails(resp, err)
	details := fmt.Sprintf(msg, args...)
	return assert.NoErrorf(t, apierr, "Error calling PDS API: %s", details)
}

func RequireNoError(t tests.T, resp *http.Response, err error) {
	t.Helper()
	if NoError(t, resp, err) {
		return
	}
	t.FailNow()
}

func RequireNoErrorf(t tests.T, resp *http.Response, err error, msg string, args ...any) {
	t.Helper()
	if NoErrorf(t, resp, err, msg, args...) {
		return
	}
	t.FailNow()
}

func RequireErrorWithStatus(t tests.T, resp *http.Response, err error, expectedStatus uint) {
	t.Helper()
	if assert.Error(t, err, "PDS API call returned no error when expected to do so.") ||
		assert.NotNil(t, resp, "Received empty response when expecting error status.") ||
		assert.Equal(t, expectedStatus, resp.StatusCode, "Received status code is different than expected.") {
		return
	}
	t.FailNow()
}

func RequireNoErrorWithStatus(t tests.T, resp *http.Response, err error, expectedStatus uint) {
	t.Helper()
	if NoError(t, resp, err) ||
		assert.NotNil(t, resp, "Received empty response.") ||
		assert.Equal(t, expectedStatus, resp.StatusCode, "Received status code is different than expected.") {
		return
	}
	t.FailNow()
}
