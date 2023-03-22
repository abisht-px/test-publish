package test

import (
	"flag"
	"fmt"
	"testing"

	"github.com/portworx/pds-integration-test/internal/api"
)

var (
	skipBackups = flag.Bool("skip-backups", false, "Skip tests related to backups.")
)

func (s *PDSTestSuite) TestDataService_WriteData() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dbPostgres,
			ImageVersionTag: "14.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dbConsul,
			ImageVersionTag: "1.14.0",
			NodeCount:       1,
		},
		{
			DataServiceName: dbCassandra,
			ImageVersionTag: "4.0.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dbRedis,
			ImageVersionTag: "7.0.5",
			NodeCount:       1,
		},
		{
			DataServiceName: dbZooKeeper,
			ImageVersionTag: "3.7.1",
			NodeCount:       3,
		},
		{
			DataServiceName: dbZooKeeper,
			ImageVersionTag: "3.8.0",
			NodeCount:       3,
		},
		{
			DataServiceName: dbKafka,
			ImageVersionTag: "3.1.1",
			NodeCount:       1,
		},
		{
			DataServiceName: dbKafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dbRabbitMQ,
			ImageVersionTag: "3.10.9",
			NodeCount:       1,
		},
		{
			DataServiceName: dbMySQL,
			ImageVersionTag: "8.0.31",
			NodeCount:       1,
		},
		{
			DataServiceName: dbMongoDB,
			ImageVersionTag: "6.0.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dbElasticSearch,
			ImageVersionTag: "8.5.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dbCouchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       1,
		},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("write-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("write-%s-n%d-", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.mustEnsureDeploymentInitialized(t, deploymentID)
			s.mustEnsureStatefulSetReady(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)

			s.mustRunBasicSmokeTest(t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_Backup() {
	if *skipBackups {
		s.T().Skip("Backup tests skipped.")
	}
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dbPostgres,
			ImageVersionTag: "11.18",
			NodeCount:       1,
		},
		{
			DataServiceName: dbPostgres,
			ImageVersionTag: "12.13",
			NodeCount:       1,
		},
		{
			DataServiceName: dbPostgres,
			ImageVersionTag: "13.9",
			NodeCount:       1,
		},
		{
			DataServiceName: dbPostgres,
			ImageVersionTag: "14.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dbCassandra,
			ImageVersionTag: "4.0.6",
			NodeCount:       1,
		},
		{
			DataServiceName: dbConsul,
			ImageVersionTag: "1.14.0",
			NodeCount:       1,
		},
		{
			DataServiceName: dbRedis,
			ImageVersionTag: "7.0.5",
			NodeCount:       1,
		},
		{
			DataServiceName: dbMySQL,
			ImageVersionTag: "8.0.31",
			NodeCount:       1,
		},
		{
			DataServiceName: dbMongoDB,
			ImageVersionTag: "6.0.3",
			NodeCount:       1,
		},
		{
			DataServiceName: dbElasticSearch,
			ImageVersionTag: "8.5.2",
			NodeCount:       1,
		},
		{
			DataServiceName: dbCouchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       1,
		},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("backup-%s-%s", deployment.DataServiceName, deployment.ImageVersionString()), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.ImageVersionString())
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.mustEnsureDeploymentInitialized(t, deploymentID)
			s.mustEnsureStatefulSetReady(t, deploymentID)

			name := generateRandomName("backup-creds")
			backupTargetConfig := s.config.backupTarget
			s3Creds := backupTargetConfig.credentials.s3
			backupCredentials := s.mustCreateS3BackupCredentials(t, s3Creds, name)
			t.Cleanup(func() { s.mustDeleteBackupCredentials(t, backupCredentials.GetId()) })

			backupTarget := s.mustCreateS3BackupTarget(t, backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
			s.mustEnsureBackupTargetCreatedInTC(t, backupTarget.GetId(), s.controlPlane.TestPDSDeploymentTargetID)
			t.Cleanup(func() { s.mustDeleteBackupTarget(t, backupTarget.GetId()) })

			backup := s.mustCreateBackup(t, deploymentID, backupTarget.GetId())
			s.mustEnsureBackupSuccessful(t, deploymentID, backup.GetClusterResourceName())
			t.Cleanup(func() { s.mustDeleteBackup(t, backup.GetId()) })
		})
	}
}

func (s *PDSTestSuite) TestDataService_UpdateImage() {
	testCases := []struct {
		spec           api.ShortDeploymentSpec
		targetVersions []string
	}{
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbMongoDB,
				ImageVersionTag: "6.0.2",
				NodeCount:       1,
			},
			targetVersions: []string{"6.0.3"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "11.16",
				NodeCount:       1,
			},
			targetVersions: []string{"11.18"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "12.11",
				NodeCount:       1,
			},
			targetVersions: []string{"12.13"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "13.7",
				NodeCount:       1,
			},
			targetVersions: []string{"13.9"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "14.2",
				NodeCount:       1,
			},
			targetVersions: []string{"14.6"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "14.4",
				NodeCount:       1,
			},
			targetVersions: []string{"14.6"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "14.5",
				NodeCount:       1,
			},
			targetVersions: []string{"14.6"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbCassandra,
				ImageVersionTag: "4.0.4",
				NodeCount:       1,
			},
			targetVersions: []string{"4.0.6"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbCassandra,
				ImageVersionTag: "4.0.5",
				NodeCount:       1,
			},
			targetVersions: []string{"4.0.6"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbRedis,
				ImageVersionTag: "7.0.0",
				NodeCount:       1,
			},
			targetVersions: []string{"7.0.5"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbRedis,
				ImageVersionTag: "7.0.2",
				NodeCount:       1,
			},
			targetVersions: []string{"7.0.5"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbRedis,
				ImageVersionTag: "7.0.4",
				NodeCount:       1,
			},
			targetVersions: []string{"7.0.5"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbKafka,
				ImageVersionTag: "3.2.0",
				NodeCount:       1,
			},
			targetVersions: []string{"3.2.3"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbKafka,
				ImageVersionTag: "3.2.1",
				NodeCount:       1,
			},
			targetVersions: []string{"3.2.3"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbRabbitMQ,
				ImageVersionTag: "3.9.21",
				NodeCount:       1,
			},
			targetVersions: []string{"3.9.22"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbRabbitMQ,
				ImageVersionTag: "3.10.6",
				NodeCount:       1,
			},
			targetVersions: []string{"3.10.9"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbRabbitMQ,
				ImageVersionTag: "3.10.7",
				NodeCount:       1,
			},
			targetVersions: []string{"3.10.9"},
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbMySQL,
				ImageVersionTag: "8.0.30",
				NodeCount:       1,
			},
			targetVersions: []string{"8.0.31"},
		},
	}

	for _, testCase := range testCases {
		for _, tvt := range testCase.targetVersions {
			tt := testCase
			targetVersionTag := tvt
			s.T().Run(fmt.Sprintf("update-%s-%s-to-%s", tt.spec.DataServiceName, tt.spec.ImageVersionString(), targetVersionTag), func(t *testing.T) {
				t.Parallel()

				tt.spec.NamePrefix = fmt.Sprintf("update-%s-", tt.spec.ImageVersionString())
				deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &tt.spec)
				t.Cleanup(func() {
					s.mustRemoveDeployment(t, deploymentID)
					s.waitForDeploymentRemoved(t, deploymentID)
				})

				// Create.
				s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
				s.mustEnsureDeploymentInitialized(t, deploymentID)
				s.mustEnsureStatefulSetReady(t, deploymentID)
				s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
				s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
				s.mustRunBasicSmokeTest(t, deploymentID)

				// Update.
				newSpec := tt.spec
				newSpec.ImageVersionTag = targetVersionTag
				s.controlPlane.MustUpdateDeployment(s.ctx, t, deploymentID, &newSpec)
				s.mustEnsureStatefulSetImage(t, deploymentID, targetVersionTag)
				s.mustEnsureStatefulSetReadyAndUpdatedReplicas(t, deploymentID)
				s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
				s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
				s.mustRunBasicSmokeTest(t, deploymentID)
			})
		}
	}
}

func (s *PDSTestSuite) TestDataService_ScaleUp() {
	testCases := []struct {
		spec    api.ShortDeploymentSpec
		scaleTo int
	}{
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "11.18",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "12.13",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "13.9",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "14.6",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbCassandra,
				ImageVersionTag: "4.0.6",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbConsul,
				ImageVersionTag: "1.14.0",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbRedis,
				ImageVersionTag: "7.0.5",
				NodeCount:       6,
			},
			scaleTo: 8,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbKafka,
				ImageVersionTag: "3.1.1",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbKafka,
				ImageVersionTag: "3.2.3",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbRabbitMQ,
				ImageVersionTag: "3.10.9",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbMySQL,
				ImageVersionTag: "8.0.31",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbMongoDB,
				ImageVersionTag: "6.0.3",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbElasticSearch,
				ImageVersionTag: "8.5.2",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbCouchbase,
				ImageVersionTag: "7.1.1",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
	}

	for _, testCase := range testCases {
		tt := testCase
		s.T().Run(fmt.Sprintf("scale-%s-%s-nodes-%v-to-%v", tt.spec.DataServiceName, tt.spec.ImageVersionString(), tt.spec.NodeCount, tt.scaleTo), func(t *testing.T) {
			t.Parallel()

			tt.spec.NamePrefix = fmt.Sprintf("scale-%s-", tt.spec.ImageVersionString())
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &tt.spec)
			t.Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})

			// Create.
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.mustEnsureDeploymentInitialized(t, deploymentID)
			s.mustEnsureStatefulSetReady(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)

			// Update.
			updateSpec := tt.spec
			updateSpec.NodeCount = tt.scaleTo
			s.controlPlane.MustUpdateDeployment(s.ctx, t, deploymentID, &updateSpec)
			s.mustEnsureStatefulSetReadyAndUpdatedReplicas(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_ScaleResources() {
	testCases := []struct {
		spec                    api.ShortDeploymentSpec
		scaleToResourceTemplate string
	}{
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "14.6",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbPostgres].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbCassandra,
				ImageVersionTag: "4.0.6",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbCassandra].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbConsul,
				ImageVersionTag: "1.14.0",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbConsul].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbKafka,
				ImageVersionTag: "3.2.3",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbKafka].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbRabbitMQ,
				ImageVersionTag: "3.10.9",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbRabbitMQ].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbMySQL,
				ImageVersionTag: "8.0.31",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbMySQL].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbMongoDB,
				ImageVersionTag: "6.0.2",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbMongoDB].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbElasticSearch,
				ImageVersionTag: "8.5.2",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbElasticSearch].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbCouchbase,
				ImageVersionTag: "7.1.1",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbCouchbase].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbRedis,
				ImageVersionTag: "7.0.5",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbRedis].ResourceTemplates[1].Name,
		},
		{
			spec: api.ShortDeploymentSpec{
				DataServiceName: dbZooKeeper,
				ImageVersionTag: "3.8.0",
				NodeCount:       3,
			},
			scaleToResourceTemplate: s.controlPlane.TestPDSTemplatesMap[dbZooKeeper].ResourceTemplates[1].Name,
		},
	}

	for _, testCase := range testCases {
		tt := testCase
		s.T().Run(fmt.Sprintf("scale-%s-%s-resources", tt.spec.DataServiceName, tt.spec.ImageVersionString()), func(t *testing.T) {
			t.Parallel()

			tt.spec.NamePrefix = fmt.Sprintf("scale-%s-", tt.spec.ImageVersionString())
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &tt.spec)
			t.Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})

			// Create.
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.mustEnsureDeploymentInitialized(t, deploymentID)
			s.mustEnsureStatefulSetReady(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)

			// Update.
			updateSpec := tt.spec
			updateSpec.ResourceSettingsTemplateName = tt.scaleToResourceTemplate
			s.controlPlane.MustUpdateDeployment(s.ctx, t, deploymentID, &updateSpec)
			s.mustEnsureStatefulSetReadyAndUpdatedReplicas(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_Recovery_FromDeletion() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dbPostgres,
			ImageVersionTag: "14.6",
			NodeCount:       3,
		},
		{
			DataServiceName: dbConsul,
			ImageVersionTag: "1.14.0",
			NodeCount:       3,
		},
		{
			DataServiceName: dbCassandra,
			ImageVersionTag: "4.0.6",
			NodeCount:       3,
		},
		{
			DataServiceName: dbRedis,
			ImageVersionTag: "7.0.5",
			NodeCount:       6,
		},
		{
			DataServiceName: dbZooKeeper,
			ImageVersionTag: "3.8.0",
			NodeCount:       3,
		},
		{
			DataServiceName: dbKafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dbRabbitMQ,
			ImageVersionTag: "3.10.9",
			NodeCount:       3,
		},
		{
			DataServiceName: dbMySQL,
			ImageVersionTag: "8.0.31",
			NodeCount:       3,
		},
		{
			DataServiceName: dbMongoDB,
			ImageVersionTag: "6.0.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dbElasticSearch,
			ImageVersionTag: "8.5.2",
			NodeCount:       3,
		},
		{
			DataServiceName: dbCouchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       3,
		},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("recover-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("recover-%s-n%d-", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.mustEnsureDeploymentInitialized(t, deploymentID)
			s.mustEnsureStatefulSetReady(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)
			//Delete pods and load test
			s.deletePods(t, deploymentID)
			s.mustEnsureStatefulSetReadyAndUpdatedReplicas(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_Metrics() {
	deployments := []api.ShortDeploymentSpec{
		{
			DataServiceName: dbCassandra,
			ImageVersionTag: "4.0.6",
			NodeCount:       3,
		},
		{
			DataServiceName: dbCouchbase,
			ImageVersionTag: "7.1.1",
			NodeCount:       3,
		},
		// TODO: https://portworx.atlassian.net/browse/DS-4878
		// {
		// 	DataServiceName: dbConsul,
		// 	ImageVersionTag: "1.14.0",
		// 	NodeCount:       3,
		// },
		{
			DataServiceName: dbKafka,
			ImageVersionTag: "3.2.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dbMongoDB,
			ImageVersionTag: "6.0.3",
			NodeCount:       3,
		},
		{
			DataServiceName: dbMySQL,
			ImageVersionTag: "8.0.31",
			NodeCount:       3,
		},
		{
			DataServiceName: dbElasticSearch,
			ImageVersionTag: "8.5.2",
			NodeCount:       3,
		},
		{
			DataServiceName: dbRabbitMQ,
			ImageVersionTag: "3.10.9",
			NodeCount:       3,
		},
		{
			DataServiceName: dbRedis,
			ImageVersionTag: "7.0.5",
			NodeCount:       6,
		},
		{
			DataServiceName: dbZooKeeper,
			ImageVersionTag: "3.8.0",
			NodeCount:       3,
		},
		{
			DataServiceName: dbPostgres,
			ImageVersionTag: "14.6",
			NodeCount:       1,
		},
	}

	for _, d := range deployments {
		deployment := d
		s.T().Run(fmt.Sprintf("metrics-%s-%s-n%d", deployment.DataServiceName, deployment.ImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("metrics-%s-n%d-", deployment.ImageVersionString(), deployment.NodeCount)
			deploymentID := s.controlPlane.MustDeployDeploymentSpec(s.ctx, t, &deployment)
			t.Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})
			s.controlPlane.MustWaitForDeploymentHealthy(s.ctx, t, deploymentID)
			s.mustEnsureDeploymentInitialized(t, deploymentID)
			s.mustEnsureStatefulSetReady(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)

			// Try to get DS metrics from prometheus.
			s.mustVerifyMetrics(t, deploymentID)
		})
	}
}
