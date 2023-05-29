package targetcluster

import (
	"context"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (tc *TargetCluster) CreateClusterIssuer(ctx context.Context, clusterIssuer *certmanagerv1.ClusterIssuer) error {
	return tc.CtrlRuntimeClient.Create(ctx, clusterIssuer)
}

func (tc *TargetCluster) GetClusterIssuer(ctx context.Context, name string) (*certmanagerv1.ClusterIssuer, error) {
	clusterIssuer := certmanagerv1.ClusterIssuer{}
	err := tc.CtrlRuntimeClient.Get(ctx, types.NamespacedName{Name: name}, &clusterIssuer)
	return &clusterIssuer, err
}

func (tc *TargetCluster) UpdateClusterIssuer(ctx context.Context, clusterIssuer *certmanagerv1.ClusterIssuer) error {
	return tc.CtrlRuntimeClient.Update(ctx, clusterIssuer)
}

func (tc *TargetCluster) DeleteClusterIssuer(ctx context.Context, clusterIssuer *certmanagerv1.ClusterIssuer) error {
	return tc.CtrlRuntimeClient.Delete(ctx, clusterIssuer)
}
