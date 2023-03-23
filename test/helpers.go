package test

import (
	"fmt"
	"testing"

	"github.com/portworx/pds-integration-test/internal/random"
)

type TestLogger struct {
	t *testing.T
}

func (l *TestLogger) Print(v ...interface{}) {
	l.t.Log(v...)
}

func (l *TestLogger) Printf(format string, v ...interface{}) {
	l.t.Logf(format, v...)
}

func shouldInstallPDSHelmChart(versionConstraints string) bool {
	return versionConstraints != "0"
}

func generateRandomName(prefix string) string {
	nameSuffix := random.AlphaNumericString(random.NameSuffixLength)
	return fmt.Sprintf("%s-integration-test-s3-%s", prefix, nameSuffix)
}
