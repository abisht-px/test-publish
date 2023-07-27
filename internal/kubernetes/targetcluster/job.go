package targetcluster

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (tc *TargetCluster) MustWaitForJobSuccess(ctx context.Context, t tests.T, namespace, jobName string) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		job, err := tc.GetJob(ctx, namespace, jobName)
		require.NoErrorf(t, err, "Getting %s/%s job from target cluster.", namespace, jobName)
		require.Truef(t, job.Status.Succeeded > 0,
			"Job did not succeed (Succeeded: %d, Failed: %d)", job.Status.Succeeded, job.Status.Failed,
		)
	})
}

func (tc *TargetCluster) MustWaitForJobFailure(ctx context.Context, t tests.T, namespace, jobName string) {
	wait.For(t, wait.StandardTimeout, wait.RetryInterval, func(t tests.T) {
		job, err := tc.GetJob(ctx, namespace, jobName)
		require.NoErrorf(t, err, "Getting %s/%s job from target cluster.", namespace, jobName)
		require.Truef(t, job.Status.Failed > 0,
			"Job did not fail (Succeeded: %d, Failed: %d)", job.Status.Succeeded, job.Status.Failed,
		)
	})
}

func (tc *TargetCluster) JobLogsMustNotContain(ctx context.Context, t *testing.T, namespace, jobName, rePattern string, since time.Time) {
	logs, err := tc.GetJobLogs(ctx, namespace, jobName, since)
	require.NoError(t, err)
	re, err := regexp.Compile(rePattern)
	require.NoErrorf(t, err, "Invalid log rexeg pattern %q.", rePattern)
	require.Nil(t, re.FindStringIndex(logs), "Job log '%s' contains pattern '%s':\n%s", jobName, rePattern, logs)
}

func (tc *TargetCluster) MustRunHostCheckJob(ctx context.Context, t tests.T, namespace string, jobNamePrefix, jobNameSuffix string, hosts, dnsIPs []string) string {
	jobName := fmt.Sprintf("%s-hostcheck-%s", jobNamePrefix, jobNameSuffix)
	image := "portworx/dnsutils"
	env := []corev1.EnvVar{{
		Name:  "HOSTS",
		Value: strings.Join(hosts, " "),
	}, {
		Name:  "DNS_IPS",
		Value: strings.Join(dnsIPs, " "),
	}}
	cmd := []string{
		"/bin/bash",
		"-c",
		"for D in $DNS_IPS; do echo \"Checking on DNS $D:\"; for H in $HOSTS; do IP=$(dig +short @$D $H 2>/dev/null | head -n1); if [ -z \"$IP\" ]; then echo \"  $H - MISSING IP\";  exit 1; else echo \"  $H $IP - OK\"; fi; done; done",
	}

	ttlSecondsAfterFinished := pointer.Int32(30)
	backOffLimit := pointer.Int32(0)
	job, err := tc.CreateJob(ctx, namespace, jobName, image, env, cmd, ttlSecondsAfterFinished, backOffLimit)
	require.NoErrorf(t, err, "Creating job %s/%s on target cluster.", namespace, jobName)
	return job.GetName()
}
