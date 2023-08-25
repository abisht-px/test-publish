package test

import (
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/pointer"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/kubernetes/targetcluster"
	"github.com/portworx/pds-integration-test/internal/wait"
)

// Update these expected values with every new release of PDS chart.
var expectedDefaultCapabilities = pds.ModelsDeploymentTargetCapabilities{
	Backup:         pointer.String("v2"),
	Capabilities:   pointer.String("v1"),
	Cassandra:      pointer.String("v2"),
	Consul:         pointer.String("v2"),
	Couchbase:      pointer.String("v2"),
	CrdReporting:   pointer.String("v4"),
	DataServiceTls: pointer.String(""),
	Database:       pointer.String("v1"),
	Elasticsearch:  pointer.String("v2"),
	EventReporting: pointer.String("v1"),
	Kafka:          pointer.String("v2"),
	Mongodb:        pointer.String("v2"),
	Mysql:          pointer.String("v2"),
	Postgresql:     pointer.String("v2"),
	Rabbitmq:       pointer.String("v2"),
	Redis:          pointer.String("v2"),
	Restore:        pointer.String("v1"),
	Sqlserver:      pointer.String("v2"),
	Zookeeper:      pointer.String("v2"),
}

func (s *PDSTestSuite) TestCapabilities_DeleteDeploymentOperator() {
	// Check that all operators' capabilities are exported to configmaps and reported to control plane.
	s.MustWaitForOperatorsToExportCapabilities()
	s.MustWaitForReportedCapabilities(expectedDefaultCapabilities)

	// Delete deployment operator.
	deploymentOperator := targetcluster.PDSOperators["deployment"]
	err := s.targetCluster.DeleteDeployment(s.ctx, targetcluster.PDSChartNamespace, deploymentOperator.Deployment)
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		// Run helm upgrade to install the deployment again.
		err := s.targetCluster.UpgradePDSChart(s.ctx)
		s.Require().NoError(err)
	})
	_, err = s.targetCluster.GetDeployment(s.ctx, targetcluster.PDSChartNamespace, deploymentOperator.Deployment)
	s.Require().True(k8serrors.IsNotFound(err))
	_, err = s.targetCluster.GetOperatorCapabilities(s.ctx, targetcluster.PDSChartNamespace, deploymentOperator.Name)
	s.Require().True(k8serrors.IsNotFound(err))

	// Verify that the updated list of capabilities has been reported to control plane.
	expectedCapabilities := pds.ModelsDeploymentTargetCapabilities{
		Backup:         pointer.String("v2"),
		Capabilities:   pointer.String(""),
		Cassandra:      pointer.String(""),
		Consul:         pointer.String(""),
		Couchbase:      pointer.String(""),
		CrdReporting:   pointer.String("v4"),
		DataServiceTls: pointer.String(""),
		Database:       pointer.String(""),
		Elasticsearch:  pointer.String(""),
		EventReporting: pointer.String("v1"),
		Kafka:          pointer.String(""),
		Mongodb:        pointer.String(""),
		Mysql:          pointer.String(""),
		Postgresql:     pointer.String(""),
		Rabbitmq:       pointer.String(""),
		Redis:          pointer.String(""),
		Restore:        pointer.String(""),
		Sqlserver:      pointer.String(""),
		Zookeeper:      pointer.String(""),
	}
	s.MustWaitForReportedCapabilities(expectedCapabilities)
}

func (s *PDSTestSuite) TestCapabilities_UpdateDataServiceTLSFeatureFlag() {
	// Check that all operators' capabilities are exported to configmaps and reported to control plane.
	s.MustWaitForOperatorsToExportCapabilities()
	s.MustWaitForReportedCapabilities(expectedDefaultCapabilities)

	// Enable the TLS feature flag.
	s.targetCluster.PDSChartConfig.DataServiceTLSEnabled = true
	err := s.targetCluster.UpgradePDSChart(s.ctx)
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		// Run helm upgrade to install the deployment again.
		s.targetCluster.PDSChartConfig.DataServiceTLSEnabled = false
		err := s.targetCluster.UpgradePDSChart(s.ctx)
		s.Require().NoError(err)
	})

	// Verify that the updated list of capabilities has been reported to control plane.
	expectedCapabilities := expectedDefaultCapabilities
	expectedCapabilities.DataServiceTls = pointer.String("v1")
	s.MustWaitForReportedCapabilities(expectedCapabilities)
}

func (s *PDSTestSuite) TestCapabilities_UninstallPDSChart() {
	// Check that all operators' capabilities are exported to configmaps and reported to control plane.
	s.MustWaitForOperatorsToExportCapabilities()
	s.MustWaitForReportedCapabilities(expectedDefaultCapabilities)

	// Verify that the default capabilities are reported to control plane.
	deploymentTarget := s.controlPlane.MustGetDeploymentTarget(s.ctx, s.T())
	reportedCapabilities := deploymentTarget.Capabilities
	s.Require().NotNil(reportedCapabilities)
	s.Require().Equal(expectedDefaultCapabilities, *reportedCapabilities)

	err := s.targetCluster.UninstallPDSChart(s.ctx)
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		err := s.targetCluster.InstallPDSChart(s.ctx)
		s.Require().NoError(err)
	})
	// Check that all operators and configmaps with capabilities were deleted.
	for _, operator := range targetcluster.PDSOperators {
		_, err := s.targetCluster.GetDeployment(s.ctx, targetcluster.PDSChartNamespace, operator.Deployment)
		s.Require().True(k8serrors.IsNotFound(err))
		_, err = s.targetCluster.GetOperatorCapabilities(s.ctx, targetcluster.PDSChartNamespace, operator.Name)
		s.Require().True(k8serrors.IsNotFound(err))
	}
}

// MustWaitForOperatorsToExportCapabilities waits until all the PDS chart operators are running
// and their configmaps with capabilities are created.
func (s *PDSTestSuite) MustWaitForOperatorsToExportCapabilities() {
	s.T().Helper()
	s.Require().EventuallyWithT(func(t *assert.CollectT) {
		for _, operator := range targetcluster.PDSOperators {
			deployment, err := s.targetCluster.GetDeployment(s.ctx, targetcluster.PDSChartNamespace, operator.Deployment)
			assert.NoError(t, err)
			assert.Equal(t, *deployment.Spec.Replicas, deployment.Status.AvailableReplicas)
			capabilities, err := s.targetCluster.GetOperatorCapabilities(s.ctx, targetcluster.PDSChartNamespace, operator.Name)
			assert.NoError(t, err)
			assert.NotEmpty(t, capabilities)
		}
	}, wait.StandardTimeout, wait.RetryInterval)
}

// MustWaitForReportedCapabilities waits until the control plane is aware of the capabilities reported by the target cluster.
func (s *PDSTestSuite) MustWaitForReportedCapabilities(expectedCapabilities pds.ModelsDeploymentTargetCapabilities) {
	s.T().Helper()
	s.Require().EventuallyWithT(func(t *assert.CollectT) {
		deploymentTarget := s.controlPlane.MustGetDeploymentTarget(s.ctx, s.T())
		reportedCapabilities := deploymentTarget.Capabilities
		assert.NotNil(t, reportedCapabilities)
		assert.Equal(t, expectedCapabilities, *reportedCapabilities)
	}, wait.ShortTimeout, wait.RetryInterval)
}
