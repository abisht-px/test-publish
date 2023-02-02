package helminstaller

import (
	"fmt"
	"strings"
)

type pdsChartConfig struct {
	VersionConstraints string
	helmChartVals      map[string]string
}

// NewPDSChartConfig stores configuration that's necessary to select certain PDS chart version and fill it with required target cluster config values.
func NewPDSChartConfig(versionConstraints, tenantID, bearerToken, APIEndpoint, clusterName string) *pdsChartConfig {
	return &pdsChartConfig{
		VersionConstraints: versionConstraints,
		helmChartVals: map[string]string{
			"tenantId":    tenantID,
			"bearerToken": bearerToken,
			"apiEndpoint": APIEndpoint,
			"clusterName": clusterName,
		},
	}
}

func (s *pdsChartConfig) CommaSeparatedChartVals() string {
	var keyValues []string

	for key, value := range s.helmChartVals {
		keyValues = append(keyValues, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(keyValues, ",")
}
