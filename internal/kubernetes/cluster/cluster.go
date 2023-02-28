package cluster

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/external-dns/endpoint"

	"github.com/portworx/pds-integration-test/internal/portforward"
)

const (
	jobNameLabel = "job-name"
)

type Cluster struct {
	config     *rest.Config
	Clientset  *kubernetes.Clientset
	MetaClient metadata.Interface
}

type ComponentSelector struct {
	Namespace     string
	LabelSelector string
}

func (s ComponentSelector) String() string {
	return fmt.Sprintf("%s/%s", s.Namespace, s.LabelSelector)
}

func NewCluster(config *rest.Config, clientset *kubernetes.Clientset, metaClient metadata.Interface) (*Cluster, error) {
	return &Cluster{
		config:     config,
		Clientset:  clientset,
		MetaClient: metaClient,
	}, nil
}

func (c *Cluster) PortforwardPod(namespace, name string, port int) (*portforward.Tunnel, error) {
	return portforward.New(c.Clientset, c.config, namespace, name, port)
}

func (c *Cluster) GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error) {
	return c.Clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Cluster) GetJob(ctx context.Context, namespace, name string) (*batchv1.Job, error) {
	return c.Clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Cluster) GetStatefulSet(ctx context.Context, namespace, name string) (*appsv1.StatefulSet, error) {
	return c.Clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Cluster) ListPods(ctx context.Context, namespace string, labelSelector map[string]string) (*corev1.PodList, error) {
	return c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(labelSelector),
	})
}

func (c *Cluster) ListServices(ctx context.Context, namespace string, labelSelector map[string]string) (*corev1.ServiceList, error) {
	return c.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(labelSelector),
	})
}

func (c *Cluster) ListDeployments(ctx context.Context, namespace string, labelSelector map[string]string) (*appsv1.DeploymentList, error) {
	return c.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(labelSelector),
	})
}

func (c *Cluster) DeletePodsBySelector(ctx context.Context, namespace string, labelSelector map[string]string) error {
	return c.Clientset.CoreV1().Pods(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(labelSelector),
	})
}

func (c *Cluster) GetPDSBackup(ctx context.Context, namespace, name string) (*backupsv1.Backup, error) {
	result := &backupsv1.Backup{}
	path := fmt.Sprintf("apis/backups.pds.io/v1/namespaces/%s/backups/%s", namespace, name)
	err := c.Clientset.RESTClient().Get().AbsPath(path).Do(ctx).Into(result)
	return result, err
}

func (c *Cluster) GetDNSEndpoints(ctx context.Context, namespace, nameFilter string, recordTypeFilter string) ([]string, error) {
	// Query DNSEndpoints.
	endpointList := &endpoint.DNSEndpointList{}
	err := c.Clientset.RESTClient().Get().
		AbsPath(fmt.Sprintf("apis/externaldns.k8s.io/v1alpha1/namespaces/%s/dnsendpoints", namespace)).
		Param("labelSelector", fmt.Sprintf("name=%s", nameFilter)).
		Do(ctx).
		Into(endpointList)
	if err != nil {
		return nil, err
	}

	// Collect DNS names.
	var dnsNames []string
	for _, item := range endpointList.Items {
		if len(item.Spec.Endpoints) > 0 {
			for _, e := range item.Spec.Endpoints {
				if len(recordTypeFilter) > 0 && e.RecordType != recordTypeFilter {
					// Skip endpoint with different record type.
					continue
				}
				dnsNames = append(dnsNames, e.DNSName)
			}
		}
	}
	return dnsNames, nil
}

func (c *Cluster) CreateJob(ctx context.Context, namespace, jobName, image string, env []corev1.EnvVar, command []string) (*batchv1.Job, error) {
	jobs := c.Clientset.BatchV1().Jobs(namespace)
	var backOffLimit int32 = 0
	var ttlSecondsAfterFinished int32 = 30

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "main",
							Image:   image,
							Env:     env,
							Command: command,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		},
	}

	job, err := jobs.Create(ctx, jobSpec, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job %s: %v", jobName, err)
	}
	return job, nil
}

func (c *Cluster) GetJobLogs(t *testing.T, ctx context.Context, namespace, jobName string, since time.Time) (string, error) {
	pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(map[string]string{
			jobNameLabel: jobName,
		}),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get pod for job '%s': %w", jobName, err)
	}
	podCount := len(pods.Items)
	if podCount == 0 {
		return "", fmt.Errorf("no pod found for job '%s'", jobName)
	}
	var logs []string
	for _, pod := range pods.Items {
		podLogs := c.getPodLogs(t, ctx, pod, since)
		if len(podLogs) > 0 {
			logs = append(logs, podLogs)
		}
	}

	return strings.Join(logs[:], "\n--------\n"), nil
}

func (c *Cluster) GetLogsForComponents(t *testing.T, ctx context.Context, components []ComponentSelector, since time.Time) {
	t.Helper()
	for _, component := range components {
		c.getComponentLogs(t, ctx, component, since)
	}
}

func (c *Cluster) getComponentLogs(t *testing.T, ctx context.Context, component ComponentSelector, since time.Time) {
	t.Helper()
	opts := metav1.ListOptions{
		LabelSelector: component.LabelSelector,
	}
	podList, err := c.Clientset.CoreV1().Pods(component.Namespace).List(ctx, opts)
	require.NoErrorf(t, err, "Listing deployment pods.")
	for _, pod := range podList.Items {
		podLogs := c.getPodLogs(t, ctx, pod, since)
		t.Logf("%s/%s:", pod.Namespace, pod.Name)
		t.Log(podLogs)
	}
}

func (c *Cluster) getPodLogs(t *testing.T, ctx context.Context, pod corev1.Pod, since time.Time) string {
	t.Helper()
	metaSince := metav1.NewTime(since)
	logOpts := &corev1.PodLogOptions{
		SinceTime: &metaSince,
	}
	req := c.Clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOpts)
	podLogs, err := req.Stream(ctx)
	require.NoError(t, err, "Streaming pod logs.")
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	require.NoError(t, err, "Copying logs buffer.")
	return buf.String()
}
