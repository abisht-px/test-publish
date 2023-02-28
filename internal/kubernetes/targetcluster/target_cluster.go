package targetcluster

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/portworx/pds-integration-test/internal/kubernetes/cluster"
	"github.com/portworx/pds-integration-test/internal/portworx"
)

const (
	pdsEnvironmentLabel = "pds/environment"
	pdsSystemNamespace  = "pds-system"
)

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

// LogComponents extracts the logs of all relevant PDS components, beginning at the specified time.
func (tc *TargetCluster) LogComponents(t *testing.T, ctx context.Context, since time.Time) {
	t.Helper()
	components := []cluster.ComponentSelector{
		{Namespace: pdsSystemNamespace, LabelSelector: "app=pds-agent"},
		// TODO (fmilichovsky): Fix log extraction
		// (the operator pods consist of two containers, so this isn't enough to qualify the one we need).
		{Namespace: pdsSystemNamespace, LabelSelector: "control-plane=controller-manager"}, // Deployment + Backup operators.
	}
	t.Log("Target cluster:")
	tc.GetLogsForComponents(t, ctx, components, since)
}

// RemoveNamespaceFinalizers removes all finalizers from a namespace.
func (tc *TargetCluster) RemoveNamespaceFinalizers(ctx context.Context, name string) (*corev1.Namespace, error) {
	namespace, err := tc.GetNamespace(ctx, name)
	if err != nil {
		return nil, err
	}
	namespace.Finalizers = []string{}
	return tc.UpdateNamespace(ctx, namespace)
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
