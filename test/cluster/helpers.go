package cluster

var (
	k8sRequiredNamespaceLabels = map[string]string{
		"pds.portworx.com/available": "true",
	}
)
