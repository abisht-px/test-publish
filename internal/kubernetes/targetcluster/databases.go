package targetcluster

import (
	"context"

	deploymentsv1 "github.com/portworx/pds-operator-deployments/api/v1"
	"github.com/stretchr/testify/require"

	"github.com/portworx/pds-integration-test/internal/tests"
	"github.com/portworx/pds-integration-test/internal/wait"
)

func (tc *TargetCluster) MustWaitForDatabaseModeNormal(ctx context.Context, t tests.T, namespace, name string) {
	wait.For(t, wait.LongTimeout, wait.StandardTimeout, func(t tests.T) {
		db, err := tc.GetPDSDatabase(ctx, namespace, name)
		require.NoErrorf(t, err, "Getting %s/%s database from target cluster.", namespace, name)
		require.Equalf(t, deploymentsv1.PDSModeNormal, db.Spec.Mode, "Database %s/%s", namespace, name)
	})
}
