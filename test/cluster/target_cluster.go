package cluster

import (
	"context"
	"testing"

	"github.anim.dreamworks.com/DreamCloud/stella-api/api/models"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	client "github.com/portworx/pds-integration-test/test/client"
	"github.com/portworx/pds-integration-test/test/color"
)

type Target struct {
	*cluster
	Token     string
	Clientset *kubernetes.Clientset
	Model     *models.TargetCluster
}

func NewTarget(token, kubeconfig string, api *client.API, clientset *kubernetes.Clientset) *Target {
	return &Target{
		cluster: &cluster{
			API:        api,
			kubeconfig: kubeconfig,
		},
		Token:     token,
		Clientset: clientset,
	}
}

func (tc *Target) MustListNodes(t *testing.T) {
	t.Helper()

	nodes, err := tc.Clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	require.NoError(t, err, "Listing all target cluster nodes.")
	require.Greater(t, len(nodes.Items), 0, "Target cluster must have at least 1 node.")
	t.Log("Target cluster nodes:")
	t.Logf("%#v", nodes.Items)
}

func (tc *Target) LogStatus(t *testing.T, namespace string) {
	t.Helper()

	t.Log(color.Blue("Target cluster:"))
	// Describe operator namespaces.
	tc.describePods(t, "pds-deployment-system")
	tc.describePods(t, "pds-backup-system")
	// Describe test namespace.
	tc.describePods(t, namespace)

	//TODO: Get logs from operators.
}
