package targetcluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/portworx/pds-integration-test/internal/kubernetes/cluster"
	"github.com/portworx/pds-integration-test/internal/portworx"
	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	pdsEnvironmentLabel = "pds/environment"
	pdsSystemNamespace  = "pds-system"

	waiterShortRetryInterval      = time.Second * 1
	waiterCoreDNSRestartedTimeout = time.Second * 30
)

var pdsUserInRedisIntroducedAt = time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC)

// TargetCluster wraps a PDS target cluster.
type TargetCluster struct {
	*cluster.Cluster
	portworx.Portworx
}

// NewTargetCluster creates a TargetCluster instance with the specified kubeconfig.
// Fails if a kubernetes go-client cannot be configured based on the kubeconfig.
func NewTargetCluster(ctx context.Context, kubeconfig string) (*TargetCluster, error) {
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

	pxNamespace, err := portworx.FindPXNamespace(ctx, clientset.CoreV1())
	if err != nil {
		return nil, err
	}
	px := portworx.New(clientset.CoreV1().RESTClient(), pxNamespace)

	cluster, err := cluster.NewCluster(config, clientset, metaClient)
	if err != nil {
		return nil, err
	}
	return &TargetCluster{
		Cluster:  cluster,
		Portworx: px,
	}, nil
}

// DeleteCRDs deletes all pds in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeleteCRDs(ctx context.Context) error {
	listOptions := metav1.ListOptions{}
	crdGroupVersionResource := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
	crdList, err := tc.MetaClient.Resource(crdGroupVersionResource).List(ctx, listOptions)
	for _, crd := range crdList.Items {
		if strings.HasSuffix(crd.Name, "pds.io") {
			crdDelErr := tc.MetaClient.Resource(crdGroupVersionResource).Delete(ctx, crd.Name, metav1.DeleteOptions{})
			if crdDelErr != nil {
				err = multierror.Append(err, crdDelErr)
			}
		}
	}
	return err
}

// DeleteClusterRoles deletes all TC cluster roles in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeleteClusterRoles(ctx context.Context) error {
	return tc.Clientset.RbacV1().ClusterRoles().DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: pdsEnvironmentLabel},
	)
}

// DeletePVCs deletes all TC PVCs in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeletePVCs(ctx context.Context, namespace string) error {
	return tc.Clientset.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: pdsEnvironmentLabel},
	)
}

// DeleteStorageClasses deletes all TC storage classes in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeleteStorageClasses(ctx context.Context) error {
	return tc.Clientset.StorageV1().StorageClasses().DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: pdsEnvironmentLabel},
	)
}

// DeleteReleasedPVs deletes all released TC PVs in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeleteReleasedPVs(ctx context.Context) error {
	pvs, err := tc.Clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, item := range pvs.Items {
		if item.Status.Phase == "Released" {
			item.Spec.PersistentVolumeReclaimPolicy = "Delete"
			_, updatePVErr := tc.Clientset.CoreV1().PersistentVolumes().Update(ctx, &item, metav1.UpdateOptions{})
			if updatePVErr != nil {
				err = multierror.Append(err, updatePVErr)
			}
		}
	}
	return err
}

// DeleteDetachedPXVolumes deletes all detached Portworx volumes in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeleteDetachedPXVolumes(ctx context.Context) error {
	// pxVolumesResponse is reduced volumes detail response from the Portworx API used for cleanup.
	type pxVolumesResponse struct {
		Volumes []struct {
			Volume struct {
				ID            string `json:"id"`
				AttachedState string `json:"attached_state"`
			} `json:"volume"`
		} `json:"volumes"`
	}

	var volumes pxVolumesResponse
	pxctlResult, err := tc.GetPXVolumes(ctx)
	if err != nil {
		return err
	}

	err = json.Unmarshal(pxctlResult, &volumes)
	if err != nil {
		return err
	}

	for _, volume := range volumes.Volumes {
		if volume.Volume.AttachedState == "ATTACH_STATE_INTERNAL" ||
			volume.Volume.AttachedState == "ATTACH_STATE_INTERNAL_SWITCH" {
			_, volumeDelErr := tc.DeletePXVolume(ctx, volume.Volume.ID)
			if volumeDelErr != nil {
				err = multierror.Append(err, volumeDelErr)
			}
		}
	}

	return err
}

func (tc *TargetCluster) MustDeleteDeploymentPods(ctx context.Context, t tests.T, namespace, deploymentID string) {
	selector := map[string]string{"pds/deployment-id": deploymentID}
	err := tc.DeletePodsBySelector(ctx, namespace, selector)
	require.NoError(t, err, "Cannot delete pods.")
}

func (tc *TargetCluster) MustFlushDNSCache(ctx context.Context, t tests.T) []string {
	// Restarts CoreDNS pods to flush DNS cache:
	// kubectl delete pods -l k8s-app=kube-dns -n kube-system
	namespace := "kube-system"
	selector := map[string]string{"k8s-app": "kube-dns"}
	err := tc.DeletePodsBySelector(ctx, namespace, selector)
	require.NoError(t, err, "Failed to delete CoreDNS pods")

	// Wait for CoreDNS pods to be fully restarted.
	wait.For(t, waiterCoreDNSRestartedTimeout, waiterShortRetryInterval, func(t tests.T) {
		set, err := tc.ListDeployments(ctx, namespace, selector)
		require.NoError(t, err, "Listing CoreDNS deployments from target cluster.")
		require.Len(t, set.Items, 1, "Expected a single CoreDNS deployment.")

		deployment := set.Items[0]
		replicas := deployment.Status.Replicas
		require.Equalf(t, replicas, deployment.Status.ReadyReplicas, "Not all replicas of deployment %s are ready.", deployment.ClusterName)
		require.Equalf(t, replicas, deployment.Status.UpdatedReplicas, "Not all replicas of deployment %s are updated.", deployment.ClusterName)
	})

	// Get and return new CoreDNS pod IPs.
	pods, err := tc.ListPods(ctx, namespace, selector)
	require.NoError(t, err, "Failed to get CoreDNS pods")
	var newPodIPs []string
	for _, pod := range pods.Items {
		if len(pod.Status.PodIP) > 0 && pod.Status.ContainerStatuses[0].Ready {
			newPodIPs = append(newPodIPs, pod.Status.PodIP)
		}
	}
	return newPodIPs
}

func (tc *TargetCluster) getDBPassword(ctx context.Context, namespace, deploymentName string) (string, error) {
	secretName := fmt.Sprintf("%s-creds", deploymentName)
	secret, err := tc.GetSecret(ctx, namespace, secretName)
	if err != nil {
		return "", err
	}

	return string(secret.Data["password"]), nil
}
