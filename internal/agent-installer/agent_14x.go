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
