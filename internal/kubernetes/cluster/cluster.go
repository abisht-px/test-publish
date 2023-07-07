package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	openstoragev1 "github.com/libopenstorage/operator/pkg/apis/core/v1"
	openstorage "github.com/libopenstorage/operator/pkg/client/clientset/versioned"
	backupsv1 "github.com/portworx/pds-operator-backups/api/v1"
	deploymentsv1 "github.com/portworx/pds-operator-deployments/api/v1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/external-dns/endpoint"

	"github.com/portworx/pds-integration-test/internal/portforward"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	jobNameLabel = "job-name"
)

type Cluster struct {
	config            *rest.Config
	Clientset         *kubernetes.Clientset
	MetaClient        metadata.Interface
	CtrlRuntimeClient ctrlclient.Client
	OpenStorageClient *openstorage.Clientset
}

func NewCluster(
	config *rest.Config,
	clientset *kubernetes.Clientset,
	metaClient metadata.Interface,
	ctrlRuntimeClient ctrlclient.Client,
	openStorageClient *openstorage.Clientset,
) (*Cluster, error) {
	return &Cluster{
		config:            config,
		Clientset:         clientset,
		MetaClient:        metaClient,
		CtrlRuntimeClient: ctrlRuntimeClient,
		OpenStorageClient: openStorageClient,
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
	listOptions := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOptions.LabelSelector = labels.FormatLabels(labelSelector)
	}
	return c.Clientset.CoreV1().Pods(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, listOptions)
}

func (c *Cluster) GetPDSBackup(ctx context.Context, namespace, name string) (*backupsv1.Backup, error) {
	result := &backupsv1.Backup{}
	path := fmt.Sprintf("apis/backups.pds.io/v1/namespaces/%s/backups/%s", namespace, name)
	err := c.Clientset.RESTClient().Get().AbsPath(path).Do(ctx).Into(result)
	return result, err
}

func (c *Cluster) GetVolumeSnapshot(ctx context.Context, namespace, name string) (runtime.Object, error) {
	path := fmt.Sprintf("apis/volumesnapshot.external-storage.k8s.io/v1/namespaces/%s/volumesnapshots/%s", namespace, name)
	res, err := c.Clientset.RESTClient().Get().AbsPath(path).Do(ctx).Get()
	return res, err
}

func (c *Cluster) GetPDSBackupJob(ctx context.Context, namespace, name string) (*backupsv1.BackupJob, error) {
	result := &backupsv1.BackupJob{}
	path := fmt.Sprintf("apis/backups.pds.io/v1/namespaces/%s/backupjobs/%s", namespace, name)
	err := c.Clientset.RESTClient().Get().AbsPath(path).Do(ctx).Into(result)
	return result, err
}

func (c *Cluster) CreatePDSRestore(ctx context.Context, namespace, name, credentialName, snapID string) (*backupsv1.Restore, error) {
	body := &backupsv1.Restore{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "backups.pds.io/v1",
			Kind:       "Restore",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: backupsv1.RestoreSpec{
			DeploymentName:      name,
			CloudCredentialName: credentialName,
			PXCloudSnapID:       snapID,
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	result := &backupsv1.Restore{}
	path := fmt.Sprintf("apis/backups.pds.io/v1/namespaces/%s/restores", namespace)
	err = c.Clientset.RESTClient().Post().AbsPath(path).Body(raw).Do(ctx).Into(result)
	return result, err
}

func (c *Cluster) GetPDSRestore(ctx context.Context, namespace, name string) (*backupsv1.Restore, error) {
	result := &backupsv1.Restore{}
	path := fmt.Sprintf("apis/backups.pds.io/v1/namespaces/%s/restores/%s", namespace, name)
	err := c.Clientset.RESTClient().Get().AbsPath(path).Do(ctx).Into(result)
	return result, err
}

func (c *Cluster) DeletePDSRestore(ctx context.Context, namespace, name string) error {
	path := fmt.Sprintf("apis/backups.pds.io/v1/namespaces/%s/restores/%s", namespace, name)
	err := c.Clientset.RESTClient().Delete().AbsPath(path).Do(ctx).Error()
	return err
}

func (c *Cluster) DeletePDSBackup(ctx context.Context, namespace, name string) error {
	path := fmt.Sprintf("apis/backups.pds.io/v1/namespaces/%s/backups/%s", namespace, name)
	err := c.Clientset.RESTClient().Delete().AbsPath(path).Do(ctx).Error()
	return err
}

func (c *Cluster) GetPDSDeployment(ctx context.Context, namespace, database, name string) (runtime.Object, error) {
	path := fmt.Sprintf("apis/deployments.pds.io/v1/namespaces/%s/%s/%s", namespace, database, name)
	res, err := c.Clientset.RESTClient().Get().AbsPath(path).Do(ctx).Get()
	return res, err
}

func (c *Cluster) DeletePDSDeployment(ctx context.Context, namespace, database, name string) error {
	path := fmt.Sprintf("apis/deployments.pds.io/v1/namespaces/%s/%s/%s", namespace, database, name)
	err := c.Clientset.RESTClient().Delete().AbsPath(path).Do(ctx).Error()
	return err
}

func (c *Cluster) GetPDSDatabase(ctx context.Context, namespace, name string) (*deploymentsv1.Database, error) {
	result := &deploymentsv1.Database{}
	path := fmt.Sprintf("apis/deployments.pds.io/v1/namespaces/%s/databases/%s", namespace, name)
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

func (c *Cluster) CreateRoleBinding(ctx context.Context, roleBindingName string, roleName string, serviceAccountName string, namespace string) error {
	_, err := c.Clientset.RbacV1().RoleBindings(namespace).Get(context.TODO(), roleBindingName, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// RoleBinding does not exist, create it
			newBinding := &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      roleBindingName,
					Namespace: namespace,
				},
				RoleRef: rbacv1.RoleRef{
					Kind: "Role",
					Name: roleName,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      serviceAccountName,
						Namespace: namespace,
					},
				},
			}
			_, err := c.Clientset.RbacV1().RoleBindings(namespace).Create(context.TODO(), newBinding, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create RoleBinding %s: %v", roleBindingName, err)
			}
			return nil
		} else {
			// Handle other errors
			return fmt.Errorf("failed to get RoleBinding %s: %v", roleBindingName, err)
		}
	}

	// RoleBinding already exists
	return nil
}

func (c *Cluster) CreateOrUpdateRole(ctx context.Context, roleName string, namespace string, rules []rbacv1.PolicyRule) error {
	existingRole, err := c.Clientset.RbacV1().Roles(namespace).Get(context.TODO(), roleName, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Role does not exist, create it
			newRole := &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      roleName,
					Namespace: namespace,
				},
				Rules: rules,
			}
			_, err := c.Clientset.RbacV1().Roles(namespace).Create(context.TODO(), newRole, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create Role %s: %v", roleName, err)
			}
			return nil
		}
		return fmt.Errorf("failed to get Role %s: %v", roleName, err)
	}

	// Role already exists, update it
	existingRole.Rules = rules
	_, err = c.Clientset.RbacV1().Roles(namespace).Update(context.TODO(), existingRole, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update Role %s: %v", roleName, err)
	}

	return nil
}

func (c *Cluster) CreateServiceAccount(ctx context.Context, serviceAccountName string, namespace string) error {
	_, err := c.Clientset.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), serviceAccountName, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// ServiceAccount does not exist, create it
			newAccount := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceAccountName,
					Namespace: namespace,
				},
			}
			_, err := c.Clientset.CoreV1().ServiceAccounts(namespace).Create(context.TODO(), newAccount, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create ServiceAccount %s: %v", serviceAccountName, err)
			}
			return nil
		} else {
			return fmt.Errorf("failed to get ServiceAccount %s: %v", serviceAccountName, err)
		}
	}
	return nil
}

func (c *Cluster) CreateLoadTestServiceAccount(ctx context.Context, namespace string) (string, error) {

	roleName := "pds-loadgen"
	serviceAccountName := "pds-loadgen"

	rules := []rbacv1.PolicyRule{
		{
			APIGroups:     []string{"policy"},
			ResourceNames: []string{"pds-restricted"},
			Resources:     []string{"podsecuritypolicies"},
			Verbs:         []string{"use"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"get"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get", "list", "delete"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"services"},
			Verbs:     []string{"get", "list"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods/exec"},
			Verbs:     []string{"create"},
		},
	}

	err := c.CreateOrUpdateRole(ctx, roleName, namespace, rules)

	if err != nil {
		return "", err
	}

	err = c.CreateServiceAccount(ctx, serviceAccountName, namespace)

	if err != nil {
		return "", err
	}

	err = c.CreateRoleBinding(ctx, fmt.Sprintf("%s-role-binding", serviceAccountName), roleName, serviceAccountName, namespace)

	if err != nil {
		return "", err
	}

	return serviceAccountName, nil
}

func (c *Cluster) CreateJob(ctx context.Context, namespace, jobName, image string, env []corev1.EnvVar, command []string, ttlSecondsAfterFinished *int32) (*batchv1.Job, error) {
	jobs := c.Clientset.BatchV1().Jobs(namespace)
	var backOffLimit int32 = 0

	spec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "main",
				Image:   image,
				Env:     env,
				Command: command,
				SecurityContext: &corev1.SecurityContext{
					AllowPrivilegeEscalation: pointer.Bool(false),
					Capabilities: &corev1.Capabilities{
						Drop: []corev1.Capability{
							"ALL",
						},
					},
				},
			},
		},
		RestartPolicy: corev1.RestartPolicyNever,
		SecurityContext: &corev1.PodSecurityContext{
			RunAsNonRoot: pointer.Bool(true),
			RunAsUser:    pointer.Int64(1000),
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		},
	}

	if strings.Contains(jobName, "-loadtest-") {
		serviceAccountName, err := c.CreateLoadTestServiceAccount(ctx, namespace)
		if err != nil {
			return nil, err
		}
		spec.ServiceAccountName = serviceAccountName
	}

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: spec,
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: ttlSecondsAfterFinished,
		},
	}

	job, err := jobs.Create(ctx, jobSpec, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job %s: %v", jobName, err)
	}
	return job, nil
}

func (c *Cluster) DeleteJob(ctx context.Context, namespace, name string) error {
	return c.Clientset.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Cluster) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return c.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Cluster) MustDeleteBackupCustomResource(ctx context.Context, t tests.T, namespace string, customResourceName string) {

	err := c.DeletePDSBackup(ctx, namespace, customResourceName)
	require.NoError(t, err)

	wait.For(t, wait.LongTimeout, wait.RetryInterval, func(t tests.T) {
		_, err := c.GetPDSBackup(ctx, namespace, customResourceName)
		expectedError := fmt.Sprintf("backups.backups.pds.io %q not found", customResourceName)
		require.EqualError(t, err, expectedError, "backup CR is not deleted.")
	})
}

func (c *Cluster) GetJobLogs(ctx context.Context, namespace, jobName string, since time.Time) (string, error) {
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
		podLogs, err := c.GetPodLogs(ctx, &pod, since)
		if err != nil {
			return "", err
		}
		if len(podLogs) > 0 {
			logs = append(logs, podLogs)
		}
	}

	return strings.Join(logs[:], "\n--------\n"), nil
}

func (c *Cluster) GetPodLogs(ctx context.Context, pod *corev1.Pod, since time.Time) (string, error) {
	metaSince := metav1.NewTime(since)
	logOpts := &corev1.PodLogOptions{
		SinceTime: &metaSince,
	}
	req := c.Clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOpts)
	podLogs, err := req.Stream(ctx)
	defer func() { _ = podLogs.Close() }()
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	return buf.String(), err
}

func (c *Cluster) GetCSIDriver(ctx context.Context, name string) (*storagev1.CSIDriver, error) {
	return c.Clientset.StorageV1().CSIDrivers().Get(ctx, name, metav1.GetOptions{})
}

func (c *Cluster) ListStorageClusters(ctx context.Context, namespace string) (*openstoragev1.StorageClusterList, error) {
	return c.OpenStorageClient.CoreV1().StorageClusters(namespace).List(ctx, metav1.ListOptions{})
}

// FindStorageCluster finds the storage cluster singleton in the cluster.
func (c *Cluster) FindStorageCluster(ctx context.Context) (*openstoragev1.StorageCluster, error) {
	storageClusters, err := c.ListStorageClusters(ctx, "")
	if err != nil {
		return nil, err
	}
	if len(storageClusters.Items) == 0 {
		return nil, errors.New("no storage cluster found")
	}
	if len(storageClusters.Items) > 1 {
		return nil, errors.New("multiple storage clusters found")
	}
	return &storageClusters.Items[0], nil
}

func (c *Cluster) UpdateStorageCluster(ctx context.Context, storageCluster *openstoragev1.StorageCluster) (*openstoragev1.StorageCluster, error) {
	return c.OpenStorageClient.CoreV1().StorageClusters(storageCluster.Namespace).Update(ctx, storageCluster, metav1.UpdateOptions{})
}

func (c *Cluster) ListStorageNodes(ctx context.Context, namespace string) (*openstoragev1.StorageNodeList, error) {
	return c.OpenStorageClient.CoreV1().StorageNodes(namespace).List(ctx, metav1.ListOptions{})
}

func (c *Cluster) GetStorageClass(ctx context.Context, name string) (*storagev1.StorageClass, error) {
	return c.Clientset.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
}

func (c *Cluster) GetCronJob(ctx context.Context, namespace, name string) (*batchv1.CronJob, error) {
	return c.Clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Cluster) MustWaitForPDSBackupWithUpdatedSchedule(ctx context.Context, t tests.T, namespace, name, schedule string) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		backupJob, err := c.GetPDSBackup(ctx, namespace, name)
		require.NoError(t, err)
		require.Equal(t, schedule, backupJob.Spec.Schedule, "backup schedule not updated in backup cr.")
	})
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		cronjob, err := c.GetCronJob(ctx, namespace, name)
		require.NoError(t, err)
		require.Equal(t, schedule, cronjob.Spec.Schedule, "backup schedule not updated in cron job.")
	})
}
