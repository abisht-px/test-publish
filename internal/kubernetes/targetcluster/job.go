package targetcluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/portworx/pds-integration-test/internal/tests"
)

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
