package helminstaller

import (
	"fmt"

	"github.com/Masterminds/semver"
)

func filterMatchingVersions(stringConstraints string, versions []string) ([]string, error) {
	constraints, err := semver.NewConstraint(stringConstraints)
	if err != nil {
		return nil, err
	}

	var matchingVersions []string
	for _, v := range versions {
		if ver, err := semver.NewVersion(v); err == nil && constraints.Check(ver) {
			matchingVersions = append(matchingVersions, v)
		}
	}
	if len(matchingVersions) == 0 {
		return nil, fmt.Errorf("no version found to match constraints %s", stringConstraints)
	}
	return matchingVersions, nil
}
