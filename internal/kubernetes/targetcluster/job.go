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

	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (tc *TargetCluster) MustWaitForJobSuccess(ctx context.Context, t tests.T, namespace, jobName string) {
	// 1. Wait for the job to finish.
	tc.MustWaitForJobToFinish(ctx, t, namespace, jobName, wait.JobFinishedTimeout, wait.ShortRetryInterval)

	// 2. Check the result.
	job, err := tc.GetJob(ctx, namespace, jobName)
	require.NoErrorf(t, err, "Getting job %s/%s from target cluster.", namespace, jobName)
	require.Greaterf(t, job.Status.Succeeded, 0, "Job %s/%s did not succeed.", namespace, jobName)
}

func (tc *TargetCluster) MustWaitForJobToFinish(ctx context.Context, t tests.T, namespace string, jobName string, timeout time.Duration, tick time.Duration) {
	wait.For(t, timeout, tick, func(t tests.T) {
		job, err := tc.GetJob(ctx, namespace, jobName)
		require.NoErrorf(t, err, "Getting %s/%s job from target cluster.", namespace, jobName)
		require.Truef(t,
			job.Status.Succeeded > 0 || job.Status.Failed > 0,
			"Job did not finish (Succeeded: %d, Failed: %d)", job.Status.Succeeded, job.Status.Failed,
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

	job, err := tc.CreateJob(ctx, namespace, jobName, image, env, cmd)
	require.NoErrorf(t, err, "Creating job %s/%s on target cluster.", namespace, jobName)
	return job.GetName()
}
