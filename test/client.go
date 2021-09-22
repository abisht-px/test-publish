package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type requestMethod func() (*http.Response, error)

func mustReachAddress(t *testing.T, url string) {
	t.Helper()
	mustProcessRequest(t,
		func() (*http.Response, error) { return http.Get(url) },
		http.StatusOK,
		fmt.Sprintf("Couldn't reach %q.", url))
}

func mustPostJSON(t *testing.T, url string, object interface{}) {
	t.Helper()

	postContent, err := json.Marshal(object)
	require.NoErrorf(t, err, "Failed to serialize data before POSTing to %q.", url)

	resp := mustProcessRequest(t,
		func() (*http.Response, error) {
			return http.Post(url, "application/json", bytes.NewReader(postContent))
		},
		http.StatusCreated,
		fmt.Sprintf("Failed to POST data to endpoint %q.", url))

	mustUnmarshalJSON(t, resp.Body, object, url)
}

func mustGetJSON(t *testing.T, url string, target interface{}) {
	t.Helper()

	resp := mustProcessRequest(t,
		func() (*http.Response, error) { return http.Get(url) },
		http.StatusOK,
		fmt.Sprintf("Failed to GET data from endpoint %q.", url))

	mustUnmarshalJSON(t, resp.Body, target, url)
}

func mustProcessRequest(t *testing.T, request requestMethod, expectedCode int, errorMessage string) *http.Response {
	t.Helper()

	resp, err := request()
	require.NoError(t, err, fmt.Sprintf("%s %v", errorMessage, err))
	require.Equal(t, expectedCode, resp.StatusCode, fmt.Sprintf("%s %s", errorMessage, "Unexpected status code."))
	return resp
}

func mustUnmarshalJSON(t *testing.T, input io.Reader, target interface{}, endpoint string) {
	content, err := io.ReadAll(input)
	require.NoErrorf(t, err, "Failed to read response body from endpoint %q. %v", endpoint, err)

	err = json.Unmarshal(content, target)
	require.NoErrorf(t, err, "Failed to unmarshal JSON content from endpoint %q. %v", endpoint, err)
}
