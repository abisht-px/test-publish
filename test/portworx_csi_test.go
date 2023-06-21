package test

import (
	"context"
	"testing"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/pointer"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
)

func (s *PDSTestSuite) TestPortworxCSI_Enabled() {
	s.targetCluster.MustSetStorageClusterCSIEnabled(s.ctx, s.T(), true)
	_, err := s.targetCluster.GetPortworxCSIDriver(s.ctx)
	s.Require().NoError(err)

	testCases := []struct {
		description          string
		requestedProvisioner string
		expectedProvisioner  string
	}{
		{
			description:          "Auto-detect provisioner",
			requestedProvisioner: "auto",
			expectedProvisioner:  targetcluster.PortworxCSIDriverName,
		},
		{
			description:          "Portworx CSI provisioner",
			requestedProvisioner: targetcluster.PortworxCSIDriverName,
			expectedProvisioner:  targetcluster.PortworxCSIDriverName,
		},
		{
			description:          "Portworx In-tree provisioner",
			requestedProvisioner: targetcluster.PortworxInTreeDriverName,
			expectedProvisioner:  "",
		},
	}
	for _, testCase := range testCases {
		tc := testCase // Make a copy for the closure.
		s.T().Run(tc.description, func(t *testing.T) {
			t.Parallel()
			template := pds.ControllersCreateStorageOptionsTemplateRequest{
				Name:        pointer.String("CSI-enabled test: " + tc.description),
				Repl:        pointer.Int32(1),
				Secure:      pointer.Bool(false),
				Fs:          pointer.String("xfs"),
				Fg:          pointer.Bool(false),
				Provisioner: pointer.String(tc.requestedProvisioner),
			}
			templateID := s.controlPlane.MustCreateStorageOptions(s.ctx, t, template)
			t.Cleanup(func() { s.controlPlane.MustDeleteStorageOptions(s.ctx, t, templateID) })

			// Create a new deployment.
			deployment := api.ShortDeploymentSpec{
				DataServiceName:   dataservices.Postgres,
				ImageVersionTag:   "14.6",
				NodeCount:         1,
				NamePrefix:        dataservices.Postgres,
				StorageOptionName: *template.Name,
			}

			deploymentID, err := s.controlPlane.DeployDeploymentSpec(context.Background(), &deployment, s.controlPlane.TestPDSNamespaceID)
			if deploymentID != "" {
				t.Cleanup(func() {
					s.controlPlane.MustRemoveDeployment(context.Background(), t, deploymentID)
					s.controlPlane.MustWaitForDeploymentRemoved(context.Background(), t, deploymentID)
				})
			}

			if tc.expectedProvisioner == "" {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			s.controlPlane.MustWaitForDeploymentAvailable(context.Background(), t, deploymentID)

			storageClasses := s.crossCluster.MustGetStorageClassesForDeployment(context.Background(), t, deploymentID)
			for _, sc := range storageClasses {
				require.Equal(t, tc.expectedProvisioner, sc.Provisioner)
				require.Equal(t, "xfs", sc.Parameters["fs"])
			}
		})
	}
}

func (s *PDSTestSuite) TestPortworxCSI_Disabled() {
	s.T().Cleanup(func() {
		s.targetCluster.MustSetStorageClusterCSIEnabled(context.Background(), s.T(), true)
	})
	s.targetCluster.MustSetStorageClusterCSIEnabled(s.ctx, s.T(), false)
	_, err := s.targetCluster.GetPortworxCSIDriver(s.ctx)
	s.Require().Error(err)

	testCases := []struct {
		description          string
		requestedProvisioner string
		expectedProvisioner  string
	}{
		{
			description:          "Auto-detect provisioner",
			requestedProvisioner: "auto",
			expectedProvisioner:  targetcluster.PortworxInTreeDriverName,
		},
		{
			description:          "Portworx CSI provisioner",
			requestedProvisioner: targetcluster.PortworxCSIDriverName,
			expectedProvisioner:  "",
		},
		{
			description:          "Portworx In-tree provisioner",
			requestedProvisioner: targetcluster.PortworxInTreeDriverName,
			expectedProvisioner:  targetcluster.PortworxInTreeDriverName,
		},
	}
	for _, testCase := range testCases {
		tc := testCase // Make a copy for the closure.
		s.T().Run(tc.description, func(t *testing.T) {
			t.Parallel()
			template := pds.ControllersCreateStorageOptionsTemplateRequest{
				Name:        pointer.String("CSI-disabled test: " + tc.description),
				Repl:        pointer.Int32(1),
				Secure:      pointer.Bool(false),
				Fs:          pointer.String("xfs"),
				Fg:          pointer.Bool(false),
				Provisioner: pointer.String(tc.requestedProvisioner),
			}
			templateID := s.controlPlane.MustCreateStorageOptions(s.ctx, t, template)
			t.Cleanup(func() { s.controlPlane.MustDeleteStorageOptions(s.ctx, t, templateID) })

			// Create a new deployment.
			deployment := api.ShortDeploymentSpec{
				DataServiceName:   dataservices.Postgres,
				ImageVersionTag:   "14.6",
				NodeCount:         1,
				NamePrefix:        dataservices.Postgres,
				StorageOptionName: *template.Name,
			}

			deploymentID, err := s.controlPlane.DeployDeploymentSpec(context.Background(), &deployment, s.controlPlane.TestPDSNamespaceID)
			if deploymentID != "" {
				t.Cleanup(func() {
					s.controlPlane.MustRemoveDeployment(context.Background(), t, deploymentID)
					s.controlPlane.MustWaitForDeploymentRemoved(context.Background(), t, deploymentID)
				})
			}

			if tc.expectedProvisioner == "" {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			s.controlPlane.MustWaitForDeploymentAvailable(context.Background(), t, deploymentID)

			storageClasses := s.crossCluster.MustGetStorageClassesForDeployment(context.Background(), t, deploymentID)
			for _, sc := range storageClasses {
				require.Equal(t, tc.expectedProvisioner, sc.Provisioner)
				require.Equal(t, "xfs", sc.Parameters["fs"])
			}
		})
	}
}
