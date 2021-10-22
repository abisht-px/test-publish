package cluster

import (
	"fmt"
	"testing"

	client "github.com/portworx/pds-integration-test/test/client"
	"github.com/portworx/pds-integration-test/test/color"
)

type ControlPlane struct {
	*cluster
}

const (
	pdsSystemNamespace = "pds-system"
)

func NewControlPlane(api *client.API, context string) *ControlPlane {
	return &ControlPlane{
		cluster: &cluster{
			API:        api,
			kubeconfig: context,
		},
	}
}

func (cp *ControlPlane) LogStatus(t *testing.T) {
	t.Helper()

	t.Log(color.Blue("Control plane:"))

	cp.describePods(t, pdsSystemNamespace)

	t.Log(headerColor("API server logs:"))
	cp.logComponent(t, pdsSystemNamespace, "api-server")

	t.Log(headerColor("API worker logs:"))
	cp.logComponent(t, pdsSystemNamespace, "api-worker")

	t.Log(headerColor("faktory logs:"))
	cp.logComponent(t, pdsSystemNamespace, "faktory")
}

func (cp *ControlPlane) logComponent(t *testing.T, namespace, name string) {
	t.Helper()

	selector := fmt.Sprintf("component=%s", name)
	t.Log("\n" + cp.MustKubectl(t,
		"logs",
		"--namespace", namespace,
		"--selector", selector,
	))
}
