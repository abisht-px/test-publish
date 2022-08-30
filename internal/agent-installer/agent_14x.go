package agent_installer

const (
	helmPDSVersionConstraint = "~1.4"
)

func NewSelectorHelmPDS14(kubeconfig, tenantID, bearerToken, APIEndpoint string) (*selectorHelmPDS, error) {
	getter := restGetterFromKubeConfig(kubeconfig)

	return &selectorHelmPDS{
		restGetter:        getter,
		versionContraints: helmPDSVersionConstraint,
		helmChartVals: map[string]string{
			"tenantId":    tenantID,
			"bearerToken": bearerToken,
			"apiEndpoint": APIEndpoint,
		},
	}, nil
}

func NewSelectorHelmPDS14WithName(kubeconfig, tenantID, bearerToken, APIEndpoint, clusterName string) (*selectorHelmPDS, error) {
	selector, error := NewSelectorHelmPDS14(kubeconfig, tenantID, bearerToken, APIEndpoint)
	if error != nil {
		return nil, error
	}
	selector.helmChartVals["clusterName"] = clusterName
	return selector, nil
}
