package test

import (
	"fmt"
	"testing"

	batchv1 "k8s.io/api/batch/v1"

	"github.com/portworx/pds-integration-test/internal/random"
)

func isJobSucceeded(job *batchv1.Job) bool {
	return *job.Spec.Completions == job.Status.Succeeded
}

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
