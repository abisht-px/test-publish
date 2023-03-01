package targetcluster

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (tc *TargetCluster) CreateNamespace(ctx context.Context, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	return tc.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
}

func (tc *TargetCluster) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	return tc.Clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
}

func (tc *TargetCluster) UpdateNamespace(ctx context.Context, namespace *corev1.Namespace) (*corev1.Namespace, error) {
	return tc.Clientset.CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})
}

func (tc *TargetCluster) DeleteNamespace(ctx context.Context, name string) error {
	return tc.Clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
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
