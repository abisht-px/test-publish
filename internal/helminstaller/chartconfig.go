package helminstaller

import (
	"fmt"
	"strings"
)

type ChartConfig struct {
	VersionConstraints string
	chartValues        map[string]string
}

// NewPDSChartConfig stores configuration that's necessary to select certain PDS chart version and fill it with required target cluster config values.
func NewPDSChartConfig(versionConstraints, tenantID, bearerToken, APIEndpoint, clusterName string) *ChartConfig {
	return &ChartConfig{
		VersionConstraints: versionConstraints,
		chartValues: map[string]string{
			"tenantId":    tenantID,
			"bearerToken": bearerToken,
			"apiEndpoint": APIEndpoint,
			"clusterName": clusterName,
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
