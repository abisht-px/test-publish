package cluster

import (
	"fmt"
	"testing"

	client "github.com/portworx/pds-integration-test/test/client"
	"github.com/portworx/pds-integration-test/test/color"
	"github.com/portworx/pds-integration-test/test/command"
)

type cluster struct {
	API        *client.API
	kubeconfig string
}

var (
	headerColor = color.Cyan
)

func (c *cluster) MustKubectl(t *testing.T, args ...string) string {
	t.Helper()
	kubectlArgs := append([]string{"--kubeconfig", c.kubeconfig}, args...)
	cmd := command.Kubectl(kubectlArgs...)
	return command.MustRun(t, cmd)
}

func (c *cluster) describePods(t *testing.T, namespace string) {
	t.Helper()

	t.Log(headerColor(fmt.Sprintf("Pods in %s:", namespace)))
	t.Log(c.MustKubectl(t,
		"describe", "pods",
		"--namespace", namespace))
}
