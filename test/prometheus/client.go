package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type pdsProxyTransport struct {
	rt       http.RoundTripper
	tenantID string
	token    string
}

func (t *pdsProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+t.token)
	req.Header.Add("X-PDS-TenantID", t.tenantID)
	return t.rt.RoundTrip(req)
}

func newPDSProxyTransport(rt http.RoundTripper, tenantID, token string) *pdsProxyTransport {
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &pdsProxyTransport{
		rt:       rt,
		tenantID: tenantID,
		token:    token,
	}
}

// NewClient builds the prometheus client for the pds proxy.
// It sets a few additional headers required for authorization.
func NewClient(address string, tenantID, token string) (prometheusv1.API, error) {
	promClient, err := api.NewClient(api.Config{
		Address:      address,
		RoundTripper: newPDSProxyTransport(nil, tenantID, token),
	})
	if err != nil {
		return nil, err
	}
	return prometheusv1.NewAPI(promClient), nil
}
