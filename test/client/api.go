package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

type API struct {
	client  *http.Client
	BaseURL string
}

func NewAPI(baseURL string) *API {
	return &API{
		client:  http.DefaultClient,
		BaseURL: baseURL,
	}
}

func (a *API) endpoint(pathFragments ...string) string {
	return fmt.Sprintf("%s/%s", a.BaseURL, path.Join(pathFragments...))
}

func (a *API) MustReachEndpoint(t *testing.T, path string) {
	t.Helper()

	target := a.endpoint(path)
	a.mustProcessRequest(t,
		func(c *http.Client) (*http.Response, error) { 
			return c.Get(target) 
		},
		http.StatusOK,
		fmt.Sprintf("Checking reachability of %q.", target))
}

func (a* API) MustPostJSON(t *testing.T, path string, object interface{}) {
	t.Helper()

	target := a.endpoint(path)
	postContent, err := json.Marshal(object)
	require.NoErrorf(t, err, "Serializing data before POSTing to %q.", target)

	resp := a.mustProcessRequest(t,
		func(c *http.Client) (*http.Response, error) { 
			return c.Post(target, "application/json", bytes.NewReader(postContent)) 
		},
		http.StatusCreated,
		fmt.Sprintf("POST data to endpoint %q.\nData: %#v\n", target, object))

	mustUnmarshalJSON(t, resp.Body, object, target)
}

func (a *API) MustGetJSON(t *testing.T, path string, object interface{}) {
	t.Helper()

	target := a.endpoint(path)
	resp := a.mustProcessRequest(t,
		func(c *http.Client) (*http.Response, error) { 
			return c.Get(target) 
		},
		http.StatusOK,
		fmt.Sprintf("GET data from endpoint %q.", target))

	mustUnmarshalJSON(t, resp.Body, object, target)
}

func (a *API) mustProcessRequest(t *testing.T, requestFunc func(*http.Client) (*http.Response, error), expectedCode int, message string) *http.Response {
	t.Helper()

	resp, err := requestFunc(a.client)
	require.NoError(t, err, fmt.Sprintf("%s %v", message, err))
	require.Equal(t, expectedCode, resp.StatusCode, fmt.Sprintf("%s %s", message, "Unexpected status code."))
	return resp
}

func mustUnmarshalJSON(t *testing.T, input io.ReadCloser, targetObject interface{}, endpoint string) {
	t.Helper()
	defer func() {
		err := input.Close()
		require.NoErrorf(t, err, "Closing response body from endpoint %q. %v", endpoint, err)
	} ()
	
	content, err := io.ReadAll(input)
	require.NoErrorf(t, err, "Reading response body from endpoint %q. %v", endpoint, err)

	err = json.Unmarshal(content, targetObject)
	require.NoErrorf(t, err, "Unmarshalling JSON content from endpoint %q. %v", endpoint, err)
}
