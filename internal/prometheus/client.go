package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"golang.org/x/oauth2"
)

type pdsProxyTransport struct {
	tenantID    string
	tokenSource oauth2.TokenSource
}

func (t *pdsProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-PDS-TenantID", t.tenantID)
	token, err := t.tokenSource.Token()
	if err != nil {
		return nil, err
	}

	token.SetAuthHeader(req)
	return http.DefaultTransport.RoundTrip(req)
}

func newPDSProxyTransport(tenantID string, tokenSource oauth2.TokenSource) *pdsProxyTransport {
	return &pdsProxyTransport{
		tenantID:    tenantID,
		tokenSource: tokenSource,
	}
}

// NewClient builds the prometheus client for the pds proxy.
// It sets a few additional headers required for authorization.
func NewClient(address string, tenantID string, tokenSource oauth2.TokenSource) (prometheusv1.API, error) {
	promClient, err := api.NewClient(api.Config{
		Address:      address,
		RoundTripper: newPDSProxyTransport(tenantID, tokenSource),
	})
	if err != nil {
		return nil, err
	}
	return prometheusv1.NewAPI(promClient), nil
}
