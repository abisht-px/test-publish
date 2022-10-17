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
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/external-dns/endpoint"

	"github.com/portworx/pds-integration-test/internal/portforward"
)

const (
	jobNameLabel = "job-name"
)

type cluster struct {
	config     *rest.Config
	clientset  *kubernetes.Clientset
	metaClient metadata.Interface
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
	metaClient, err := metadata.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &cluster{
		config:     config,
		clientset:  clientset,
		metaClient: metaClient,
	}, nil
}

func (c *cluster) PortforwardPod(namespace, name string, port int) (*portforward.Tunnel, error) {
	return portforward.New(c.clientset, c.config, namespace, name, port)
}

func (c *cluster) GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *cluster) GetJob(ctx context.Context, namespace, name string) (*batchv1.Job, error) {
	return c.clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *cluster) GetStatefulSet(ctx context.Context, namespace, name string) (*appsv1.StatefulSet, error) {
	return c.clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *cluster) ListPods(ctx context.Context, namespace string, labelSelector map[string]string) (*corev1.PodList, error) {
	return c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(labelSelector),
	})
}

func (c *cluster) ListServices(ctx context.Context, namespace string, labelSelector map[string]string) (*corev1.ServiceList, error) {
	return c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(labelSelector),
	})
}

func (c *cluster) ListDeployments(ctx context.Context, namespace string, labelSelector map[string]string) (*appsv1.DeploymentList, error) {
	return c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(labelSelector),
	})
}

func (c *cluster) DeletePodsBySelector(ctx context.Context, namespace string, labelSelector map[string]string) error {
	return c.clientset.CoreV1().Pods(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(labelSelector),
	})
}

func (c *cluster) GetPDSBackup(ctx context.Context, namespace, name string) (*backupsv1.Backup, error) {
	result := &backupsv1.Backup{}
	path := fmt.Sprintf("apis/backups.pds.io/v1/namespaces/%s/backups/%s", namespace, name)
	err := c.clientset.RESTClient().Get().AbsPath(path).Do(ctx).Into(result)
	return result, err
}

func (c *cluster) GetDNSEndpoints(ctx context.Context, namespace, nameFilter string, recordTypeFilter string) ([]string, error) {
	// Query DNSEndpoints.
	endpointList := &endpoint.DNSEndpointList{}
	err := c.clientset.RESTClient().Get().
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

func (c *cluster) CreateJob(ctx context.Context, namespace, jobName, image string, env []corev1.EnvVar, command []string) (*batchv1.Job, error) {
	jobs := c.clientset.BatchV1().Jobs(namespace)
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

func (c *cluster) GetJobLogs(t *testing.T, ctx context.Context, namespace, jobName string, since time.Time) (string, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
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

func (c *cluster) ensurePDSNamespaceLabels(t *testing.T, ctx context.Context, namespaces []string) {
	t.Helper()
	for _, ns := range namespaces {
		k8sns, err := c.clientset.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
		if err != nil {
			k8sns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   ns,
					Labels: k8sRequiredNamespaceLabels,
				},
			}
			_, err := c.clientset.CoreV1().Namespaces().Create(ctx, k8sns, metav1.CreateOptions{})
			require.NoError(t, err, "Creating namespace %s", k8sns.Name)
			continue
		}

		updateRequired := false
		for key, requiredValue := range k8sRequiredNamespaceLabels {
			if actualValue, ok := k8sns.Labels[key]; !ok || actualValue != requiredValue {
				k8sns.Labels[key] = requiredValue
				updateRequired = true
			}
		}
		if !updateRequired {
			continue
		}

		_, err = c.clientset.CoreV1().Namespaces().Update(ctx, k8sns, metav1.UpdateOptions{})
		require.NoError(t, err, "Updating namespace %s", k8sns.Name)
	}
}
