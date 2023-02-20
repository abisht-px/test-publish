package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func NoError(t *testing.T, resp *http.Response, err error) bool {
	t.Helper()
	apierr := ExtractErrorDetails(resp, err)
	return assert.NoError(t, apierr, "Error calling PDS API.")
}

func NoErrorf(t *testing.T, resp *http.Response, err error, msg string, args ...any) bool {
	t.Helper()
	apierr := ExtractErrorDetails(resp, err)
	details := fmt.Sprintf(msg, args...)
	return assert.NoErrorf(t, apierr, "Error calling PDS API: %s", details)
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
