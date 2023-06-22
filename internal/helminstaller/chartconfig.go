package helminstaller

import (
	"fmt"
	"strings"
)

type ChartConfig struct {
	VersionConstraints string
	ReleaseName        string
	ChartValues        map[string]string
}

func (s *ChartConfig) CommaSeparatedChartVals() string {
	var keyValues []string

	for key, value := range s.ChartValues {
		keyValues = append(keyValues, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(keyValues, ",")
}
