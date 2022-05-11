package cluster

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type cluster struct {
	clientset kubernetes.Interface
}

type componentSelector struct {
	namespace     string
	labelSelector string
}

func (s componentSelector) String() string {
	return fmt.Sprintf("%s/%s", s.namespace, s.labelSelector)
}

func newCluster(kubeconfig string) (*cluster, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &cluster{
		clientset: clientset,
	}, nil
}

func (c *cluster) getLogsForComponents(t *testing.T, ctx context.Context, components []componentSelector, since time.Time) {
	t.Helper()
	for _, component := range components {
		c.getComponentLogs(t, ctx, component, since)
	}
}

func (c *cluster) getComponentLogs(t *testing.T, ctx context.Context, component componentSelector, since time.Time) {
	t.Helper()
	opts := metav1.ListOptions{
		LabelSelector: component.labelSelector,
	}
	podList, err := c.clientset.CoreV1().Pods(component.namespace).List(ctx, opts)
	require.NoErrorf(t, err, "Listing deployment pods.")
	for _, pod := range podList.Items {
		podLogs := c.getPodLogs(t, ctx, pod, since)
		t.Logf("%s/%s:", pod.Namespace, pod.Name)
		t.Log(podLogs)
	}
}

func (c *cluster) getPodLogs(t *testing.T, ctx context.Context, pod corev1.Pod, since time.Time) string {
	t.Helper()
	metaSince := metav1.NewTime(since)
	logOpts := &corev1.PodLogOptions{
		SinceTime: &metaSince,
	}
	req := c.clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOpts)
	podLogs, err := req.Stream(ctx)
	require.NoError(t, err, "Streaming pod logs.")
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	require.NoError(t, err, "Copying logs buffer.")
	return buf.String()
}
