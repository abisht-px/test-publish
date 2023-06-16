package psa

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	PSAPolicyPrivileged = "privileged"
	PSAPolicyBaseline   = "baseline"
	PSAPolicyRestricted = "restricted"
)

func NewNamespace(namespaceName, psaPolicy string, pdsAvailable bool) *corev1.Namespace {
	labels := map[string]string{
		"pod-security.kubernetes.io/enforce": psaPolicy,
		"pod-security.kubernetes.io/audit":   psaPolicy,
		"pod-security.kubernetes.io/warn":    psaPolicy,
	}
	if pdsAvailable {
		labels["pds.portworx.com/available"] = "true"
	}
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespaceName,
			Labels: labels,
		},
	}
}
