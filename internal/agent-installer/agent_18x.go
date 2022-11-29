package agent_installer

const (
	helmPDSVersionConstraint = "~1.10"
)

func NewSelectorHelmPDS18(kubeconfig, tenantID, bearerToken, APIEndpoint string) (*selectorHelmPDS, error) {
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

func NewSelectorHelmPDS18WithName(kubeconfig, tenantID, bearerToken, APIEndpoint, clusterName string) (*selectorHelmPDS, error) {
	selector, err := NewSelectorHelmPDS18(kubeconfig, tenantID, bearerToken, APIEndpoint)
	if err != nil {
		return nil, err
	}
	selector.helmChartVals["clusterName"] = clusterName
	return selector, nil
}
