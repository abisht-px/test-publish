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
