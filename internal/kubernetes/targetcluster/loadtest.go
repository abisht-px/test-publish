package targetcluster

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (tc *TargetCluster) MustWaitForLoadTestSuccess(ctx context.Context, t *testing.T, namespace, jobName string, startTime time.Time) {
	// 1. Wait for the job to finish.
	tc.MustWaitForJobToFinish(ctx, t, namespace, jobName, wait.StandardTimeout, wait.ShortRetryInterval)

	// 2. Check the result.
	job, err := tc.GetJob(ctx, namespace, jobName)
	require.NoError(t, err)

	if job.Status.Failed > 0 {
		// Job failed.
		logs, err := tc.GetJobLogs(ctx, namespace, jobName, startTime)
		if err != nil {
			require.Fail(t, fmt.Sprintf("Job '%s' failed.", jobName))
		} else {
			require.Fail(t, fmt.Sprintf("Job '%s' failed. See job logs for more details:", jobName), logs)
		}
	}
	require.Greater(t, job.Status.Succeeded, int32(0), "Job %q did not succeed.", jobName)
}

func (tc *TargetCluster) MustWaitForLoadTestFailure(ctx context.Context, t *testing.T, namespace, jobName string) {
	// 1. Wait for the job to finish.
	tc.MustWaitForJobToFinish(ctx, t, namespace, jobName, wait.StandardTimeout, wait.ShortRetryInterval)

	// 2. Check the result.
	job, err := tc.GetJob(ctx, namespace, jobName)
	require.NoError(t, err)

	require.Greater(t, job.Status.Failed, int32(0), "Job %q did not fail.", jobName)
}

func (tc *TargetCluster) MustGetLoadTestJobEnv(ctx context.Context, t *testing.T, dataServiceType, deploymentName, namespace, mode, seed, user string, nodeCount int32, extraEnv map[string]string) []corev1.EnvVar {
	host := fmt.Sprintf("%s-%s", deploymentName, namespace)
	password, err := tc.getDBPassword(ctx, namespace, deploymentName)
	require.NoErrorf(t, err, "Could not get password for database %s/%s.", namespace, deploymentName)
	env := []corev1.EnvVar{
		{
			Name:  "KIND",
			Value: dataservices.ToShortName(dataServiceType),
		},
		{
			Name:  "NAMESPACE",
			Value: namespace,
		},
		{
			Name:  "HOST",
			Value: host,
		},
		{
			Name:  "PASSWORD",
			Value: password,
		},
		{
			Name:  "ITERATIONS",
			Value: "1",
		},
		{
			Name:  "FAIL_ON_ERROR",
			Value: "true",
		},
	}
	if mode != "" {
		env = append(env, corev1.EnvVar{
			Name:  "MODE",
			Value: mode,
		})
	}
	if seed != "" {
		seed := strings.ReplaceAll(seed, "-", "")
		env = append(env, corev1.EnvVar{
			Name:  "SEED",
			Value: seed,
		})
	}

	switch dataServiceType {
	case dataservices.Redis:
		var clusterMode string
		if nodeCount > 1 {
			clusterMode = "true"
		} else {
			clusterMode = "false"
		}
		env = append(env,
			corev1.EnvVar{
				Name:  "CLUSTER_MODE",
				Value: clusterMode,
			},
		)
	}

	// Set user.
	if user == "" {
		user = "pds"
	}
	env = append(env, corev1.EnvVar{
		Name:  "PDS_USER",
		Value: user,
	})

	// Set extra env or override existing ones.
	if len(extraEnv) > 0 {
		env = mergeEnvs(env, extraEnv)
	}

	return env
}

func mergeEnvs(envs []corev1.EnvVar, extra map[string]string) []corev1.EnvVar {
	mergedEnv := make(map[string]corev1.EnvVar)
	for _, value := range envs {
		mergedEnv[value.Name] = value
	}
	for name, value := range extra {
		mergedEnv[name] = corev1.EnvVar{
			Name:  name,
			Value: value,
		}
	}
	var out []corev1.EnvVar
	for _, v := range mergedEnv {
		out = append(out, v)
	}
	return out
}
