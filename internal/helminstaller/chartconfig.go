package helminstaller

import (
	"fmt"
	"strings"
)

type ChartConfig struct {
	VersionConstraints string
	ReleaseName        string
	chartValues        map[string]string
}

// NewPDSChartConfig stores configuration that's necessary to select certain PDS chart version and fill it with required target cluster config values.
func NewPDSChartConfig(versionConstraints, tenantID, bearerToken, APIEndpoint, clusterName string) *ChartConfig {
	return &ChartConfig{
		ReleaseName:        pdsReleaseName,
		VersionConstraints: versionConstraints,
		chartValues: map[string]string{
			"tenantId":    tenantID,
			"bearerToken": bearerToken,
			"apiEndpoint": APIEndpoint,
			"clusterName": clusterName,
		},
	}
}

// NewCertManagerChartConfig stores configuration that's necessary to select certain cert-manager chart version and fill it with required config values.
func NewCertManagerChartConfig(versionConstraints string) *ChartConfig {
	return &ChartConfig{
		ReleaseName:        certManagerReleaseName,
		VersionConstraints: versionConstraints,
		chartValues: map[string]string{
			"installCRDs": "true",
		},
	}
}

func (s *ChartConfig) CommaSeparatedChartVals() string {
	var keyValues []string

	for key, value := range s.chartValues {
		keyValues = append(keyValues, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(keyValues, ",")
}
