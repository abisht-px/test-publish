package targetcluster

import (
	"context"
	"fmt"
	"testing"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/wait"
)

const (
	pdsAPITimeFormat = "2006-01-02T15:04:05.999999Z"
)

func (tc *TargetCluster) MustWaitForLoadTestSuccess(ctx context.Context, t *testing.T, namespace, jobName string, startTime time.Time) {
	// 1. Wait for the job to finish.
	tc.MustWaitForJobToFinish(ctx, t, namespace, jobName, wait.LoadTestJobFinishedTimeout, wait.ShortRetryInterval)

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
	require.Greater(t, job.Status.Succeeded, 0, "Job %q did not succeed.", jobName)
}

func (tc *TargetCluster) MustGetLoadTestJobEnv(ctx context.Context, t *testing.T, dataService *pds.ModelsDataService, dsImageCreatedAt, deploymentName, namespace string, nodeCount *int32) []corev1.EnvVar {
	host := fmt.Sprintf("%s-%s", deploymentName, namespace)
	password, err := tc.getDBPassword(ctx, namespace, deploymentName)
	require.NoErrorf(t, err, "Could not get password for database %s/%s.", namespace, deploymentName)
	env := []corev1.EnvVar{
		{
			Name:  "HOST",
			Value: host,
		}, {
			Name:  "PASSWORD",
			Value: password,
		}, {
			Name:  "ITERATIONS",
			Value: "1",
		}, {
			Name:  "FAIL_ON_ERROR",
			Value: "true",
		}}

	dataServiceType := dataService.GetName()
	switch dataServiceType {
	case dataservices.Redis:
		var clusterMode string
		if nodeCount != nil && *nodeCount > 1 {
			clusterMode = "true"
		} else {
			clusterMode = "false"
		}
		var user = "pds"
		if dsImageCreatedAt != "" {
			dsCreatedAt, err := time.Parse(pdsAPITimeFormat, dsImageCreatedAt)
			if err == nil && dsCreatedAt.Before(pdsUserInRedisIntroducedAt) {
				// Older images before this change: https://github.com/portworx/pds-images-redis/pull/61 had "default" user.
				user = "default"
			}
		}
		env = append(env,
			corev1.EnvVar{
				Name:  "PDS_USER",
				Value: user,
			},
			corev1.EnvVar{
				Name:  "CLUSTER_MODE",
				Value: clusterMode,
			},
		)
	}

	return env
}
