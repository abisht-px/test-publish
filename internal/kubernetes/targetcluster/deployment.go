package targetcluster

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (tc *TargetCluster) CreateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return tc.Clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
}

func (tc *TargetCluster) GetDeployment(ctx context.Context, namespace string, name string) (*appsv1.Deployment, error) {
	return tc.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (tc *TargetCluster) UpdateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return tc.Clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
}

func (tc *TargetCluster) DeleteDeployment(ctx context.Context, namespace string, name string) error {
	return tc.Clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
