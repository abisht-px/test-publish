package test

import (
	"fmt"

	"github.com/portworx/pds-integration-test/internal/random"
)

func shouldInstallPDSHelmChart(versionConstraints string) bool {
	return versionConstraints != "0"
}

func generateRandomName(prefix string) string {
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	return fmt.Sprintf("%s-integration-test-s3-%s", prefix, nameSuffix)
}
