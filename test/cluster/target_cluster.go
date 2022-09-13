package cluster

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

const pdsEnvironmentLabel = "pds/environment"

// TargetCluster wraps a PDS target cluster.
type TargetCluster struct {
	*cluster
}

// NewTargetCluster creates a TargetCluster instance with the specified kubeconfig.
// Fails if a kubernetes go-client cannot be configured based on the kubeconfig.
func NewTargetCluster(kubeconfig string) (*TargetCluster, error) {
	cluster, err := newCluster(kubeconfig)
	if err != nil {
		return nil, err
	}
	return &TargetCluster{cluster}, nil
}

// LogComponents extracts the logs of all relevant PDS components, beginning at the specified time.
func (tc *TargetCluster) LogComponents(t *testing.T, ctx context.Context, since time.Time) {
	t.Helper()
	components := []componentSelector{
		{pdsSystemNamespace, "app=pds-agent"},
		// TODO (fmilichovsky): Fix log extraction
		// (the operator pods consist of two containers, so this isn't enough to qualify the one we need).
		{pdsSystemNamespace, "control-plane=controller-manager"}, // Deployment + Backup operators.
	}
	t.Log("Target cluster:")
	tc.getLogsForComponents(t, ctx, components, since)
}

func (tc *TargetCluster) EnsureNamespaces(t *testing.T, ctx context.Context, namespaces []string) {
	t.Helper()
	tc.ensurePDSNamespaceLabels(t, ctx, namespaces)
}

// DeleteCRDs deletes all pds in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeleteCRDs(ctx context.Context) error {
	listOptions := metav1.ListOptions{}
	crdGroupVersionResource := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
	crdList, err := tc.metaClient.Resource(crdGroupVersionResource).List(ctx, listOptions)
	for _, crd := range crdList.Items {
		if strings.HasSuffix(crd.Name, "pds.io") {
			crdDelErr := tc.metaClient.Resource(crdGroupVersionResource).Delete(ctx, crd.Name, metav1.DeleteOptions{})
			err = multierror.Append(err, crdDelErr)
		}
	}
	return err
}

// DeleteClusterRoles deletes all TC cluster roles in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeleteClusterRoles(ctx context.Context) error {
	return tc.clientset.RbacV1().ClusterRoles().DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: pdsEnvironmentLabel},
	)
}

// DeletePVCs deletes all TC PVCs in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeletePVCs(ctx context.Context, namespace string) error {
	return tc.clientset.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: pdsEnvironmentLabel},
	)
}

// DeleteStorageClasses deletes all TC storage classes in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeleteStorageClasses(ctx context.Context) error {
	return tc.clientset.StorageV1().StorageClasses().DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: pdsEnvironmentLabel},
	)
}

// DeleteReleasedPVs deletes all released TC PVs in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeleteReleasedPVs(ctx context.Context) error {
	pvs, err := tc.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, item := range pvs.Items {
		if item.Status.Phase == "Released" {
			item.Spec.PersistentVolumeReclaimPolicy = "Delete"
			_, updatePVErr := tc.clientset.CoreV1().PersistentVolumes().Update(ctx, &item, metav1.UpdateOptions{})
			err = multierror.Append(err, updatePVErr)
		}
	}
	return err
}

// DeleteDetachedPXVolumes deletes all detached Portworx volumes in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeleteDetachedPXVolumes(ctx context.Context, pxNamespace string) error {
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
	pxctlResult, err := tc.getPxVolumes(ctx, pxNamespace)
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
			_, volumeDelErr := tc.deletePxVolume(ctx, pxNamespace, volume.Volume.ID)
			err = multierror.Append(err, volumeDelErr)
		}
	}

	return err
}

// DeletePXCredentials deletes all Portworx credentials in the target cluster. Used in the test cleanup.
func (tc *TargetCluster) DeletePXCredentials(ctx context.Context, pxNamespace string) error {
	// pxCredentialsInspectResponse is reduced credentials detail response from the Portworx API used for cleanup.
	type pxCredentialsInspectResponse struct {
		Name string `json:"name"`
	}

	// pxCredentialsResponse is credentials response from the Portworx API used for cleanup.
	type pxCredentialsResponse struct {
		CredentialIDs []string `json:"credential_ids"`
	}

	credentialsJSON, err := tc.getPXCredentials(ctx, pxNamespace)
	if err != nil {
		return err
	}

	var credentialsResponse pxCredentialsResponse
	err = json.Unmarshal(credentialsJSON, &credentialsResponse)
	if err != nil {
		return err
	}

	for _, credentialID := range credentialsResponse.CredentialIDs {
		credentialDetailJSON, getCredentialsErr := tc.getPXCredentialDetail(ctx, pxNamespace, credentialID)
		if err != nil {
			err = multierror.Append(err, getCredentialsErr)
			continue
		}
		var credentialDetail pxCredentialsInspectResponse
		unmarshalErr := json.Unmarshal(credentialDetailJSON, &credentialDetail)
		if err != nil {
			err = multierror.Append(err, unmarshalErr)
			continue
		}
		if strings.HasPrefix(credentialDetail.Name, "pdscreds-") {
			_, deleteCredentialErr := tc.deletePXCredential(ctx, pxNamespace, credentialID)
			err = multierror.Append(err, deleteCredentialErr)
		}
	}

	return nil
}

// region <<pxctl utility functions>>

func buildPxCtlRequest(pxNamespace string, baseRequest *rest.Request, pathSuffix string) *rest.Request {
	return baseRequest.Namespace(pxNamespace).
		Resource("services").
		Name("portworx-api:9021").
		SubResource("proxy").
		Suffix(pathSuffix)
}

func (tc *TargetCluster) getPxVolumes(
	ctx context.Context,
	pxNamespace string,
) ([]byte, error) {
	return buildPxCtlRequest(
		pxNamespace,
		tc.clientset.CoreV1().RESTClient().Post(),
		"v1/volumes/inspectwithfilters",
	).Do(ctx).Raw()
}

func (tc *TargetCluster) deletePxVolume(
	ctx context.Context,
	pxNamespace string,
	volumeId string,
) ([]byte, error) {
	return buildPxCtlRequest(
		pxNamespace,
		tc.clientset.CoreV1().RESTClient().Delete(),
		"v1/volumes/"+volumeId,
	).Do(ctx).Raw()
}

func (tc *TargetCluster) getPXCredentials(
	ctx context.Context,
	pxNamespace string,
) ([]byte, error) {
	return buildPxCtlRequest(
		pxNamespace,
		tc.clientset.CoreV1().RESTClient().Get(),
		"v1/credentials",
	).Do(ctx).Raw()
}

func (tc *TargetCluster) deletePXCredential(
	ctx context.Context,
	pxNamespace string,
	credentialID string,
) ([]byte, error) {
	return buildPxCtlRequest(
		pxNamespace,
		tc.clientset.CoreV1().RESTClient().Delete(),
		"v1/credentials/"+credentialID,
	).Do(ctx).Raw()
}

func (tc *TargetCluster) getPXCredentialDetail(
	ctx context.Context,
	pxNamespace string,
	credentialID string,
) ([]byte, error) {
	return buildPxCtlRequest(
		pxNamespace,
		tc.clientset.CoreV1().RESTClient().Get(),
		"v1/credentials/inspect/"+credentialID,
	).Do(ctx).Raw()
}

// endregion <<pxctl utility functions>>
