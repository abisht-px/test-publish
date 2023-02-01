package agent_installer

// NewSelectorHelmPDS enables arbitrary version selector with hardcoded tenantId, bearerToken, apiEndpoint. Use it with caution.
// There is no incode validation of the version and no guarantee that the version is available. Version validation is done when Installer is created from selector.
// There is no incode validation of the helm chart arguments. Chart arguments are validated only during installation.
func NewSelectorHelmPDS(kubeconfig, versionConstraints, tenantID, bearerToken, APIEndpoint, clusterName string) *selectorHelmPDS {
	return &selectorHelmPDS{
		restGetter:         restGetterFromKubeConfig(kubeconfig),
		versionConstraints: versionConstraints,
		helmChartVals: map[string]string{
			"tenantId":    tenantID,
			"bearerToken": bearerToken,
			"apiEndpoint": APIEndpoint,
			"clusterName": clusterName,
		},
	}
}
