package test

import (
	"flag"
	"fmt"
	"testing"
)

var (
	skipBackups = flag.Bool("skip-backups", false, "Skip tests related to backups.")
)

func (s *PDSTestSuite) TestDataService_WriteData() {
	deployments := []ShortDeploymentSpec{
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
		s.T().Run(fmt.Sprintf("write-%s-%s-n%d", deployment.DataServiceName, deployment.getImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("write-%s-n%d-", deployment.getImageVersionString(), deployment.NodeCount)
			deploymentID := s.mustDeployDeploymentSpec(t, deployment)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})
			s.mustEnsureDeploymentHealthy(t, deploymentID)
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
	deployments := []ShortDeploymentSpec{
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
		s.T().Run(fmt.Sprintf("backup-%s-%s", deployment.DataServiceName, deployment.getImageVersionString()), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("backup-%s-", deployment.getImageVersionString())
			deploymentID := s.mustDeployDeploymentSpec(t, deployment)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})
			s.mustEnsureDeploymentHealthy(t, deploymentID)
			s.mustEnsureDeploymentInitialized(t, deploymentID)
			s.mustEnsureStatefulSetReady(t, deploymentID)

			name := generateRandomName("backup-creds")
			backupTargetConfig := s.config.backupTarget
			s3Creds := backupTargetConfig.credentials.s3
			backupCredentials := s.mustCreateS3BackupCredentials(t, s3Creds, name)
			s.T().Cleanup(func() { s.mustDeleteBackupCredentials(t, backupCredentials.GetId()) })

			backupTarget := s.mustCreateS3BackupTarget(t, backupCredentials.GetId(), backupTargetConfig.bucket, backupTargetConfig.region)
			s.mustEnsureBackupTargetCreatedInTC(t, backupTarget.GetId(), s.testPDSDeploymentTargetID)
			s.T().Cleanup(func() { s.mustDeleteBackupTarget(t, backupTarget.GetId()) })

			backup := s.mustCreateBackup(t, deploymentID, backupTarget.GetId())
			s.mustEnsureBackupSuccessful(t, deploymentID, backup.GetClusterResourceName())
			s.T().Cleanup(func() { s.mustDeleteBackup(t, backup.GetId()) })
		})
	}
}

func (s *PDSTestSuite) TestDataService_UpdateImage() {
	testCases := []struct {
		spec           ShortDeploymentSpec
		targetVersions []string
	}{
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbMongoDB,
				ImageVersionTag: "6.0.2",
				NodeCount:       1,
			},
			targetVersions: []string{"6.0.3"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "11.16",
				NodeCount:       1,
			},
			targetVersions: []string{"11.18"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "12.11",
				NodeCount:       1,
			},
			targetVersions: []string{"12.13"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "13.7",
				NodeCount:       1,
			},
			targetVersions: []string{"13.9"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "14.2",
				NodeCount:       1,
			},
			targetVersions: []string{"14.6"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "14.4",
				NodeCount:       1,
			},
			targetVersions: []string{"14.6"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "14.5",
				NodeCount:       1,
			},
			targetVersions: []string{"14.6"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbCassandra,
				ImageVersionTag: "4.0.4",
				NodeCount:       1,
			},
			targetVersions: []string{"4.0.6"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbCassandra,
				ImageVersionTag: "4.0.5",
				NodeCount:       1,
			},
			targetVersions: []string{"4.0.6"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbRedis,
				ImageVersionTag: "7.0.0",
				NodeCount:       1,
			},
			targetVersions: []string{"7.0.5"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbRedis,
				ImageVersionTag: "7.0.2",
				NodeCount:       1,
			},
			targetVersions: []string{"7.0.5"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbRedis,
				ImageVersionTag: "7.0.4",
				NodeCount:       1,
			},
			targetVersions: []string{"7.0.5"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbKafka,
				ImageVersionTag: "3.2.0",
				NodeCount:       1,
			},
			targetVersions: []string{"3.2.3"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbKafka,
				ImageVersionTag: "3.2.1",
				NodeCount:       1,
			},
			targetVersions: []string{"3.2.3"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbRabbitMQ,
				ImageVersionTag: "3.9.21",
				NodeCount:       1,
			},
			targetVersions: []string{"3.9.22"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbRabbitMQ,
				ImageVersionTag: "3.10.6",
				NodeCount:       1,
			},
			targetVersions: []string{"3.10.9"},
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbRabbitMQ,
				ImageVersionTag: "3.10.7",
				NodeCount:       1,
			},
			targetVersions: []string{"3.10.9"},
		},
		{
			spec: ShortDeploymentSpec{
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
			s.T().Run(fmt.Sprintf("update-%s-%s-to-%s", tt.spec.DataServiceName, tt.spec.getImageVersionString(), targetVersionTag), func(t *testing.T) {
				t.Parallel()

				tt.spec.NamePrefix = fmt.Sprintf("update-%s-", tt.spec.getImageVersionString())
				deploymentID := s.mustDeployDeploymentSpec(t, tt.spec)
				s.T().Cleanup(func() {
					s.mustRemoveDeployment(t, deploymentID)
					s.waitForDeploymentRemoved(t, deploymentID)
				})

				// Create.
				s.mustEnsureDeploymentHealthy(t, deploymentID)
				s.mustEnsureDeploymentInitialized(t, deploymentID)
				s.mustEnsureStatefulSetReady(t, deploymentID)
				s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
				s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
				s.mustRunBasicSmokeTest(t, deploymentID)

				// Update.
				newSpec := tt.spec
				newSpec.ImageVersionTag = targetVersionTag
				s.mustUpdateDeployment(t, deploymentID, &newSpec)
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
		spec    ShortDeploymentSpec
		scaleTo int
	}{
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "11.18",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "12.13",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "13.9",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "14.6",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbCassandra,
				ImageVersionTag: "4.0.6",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbConsul,
				ImageVersionTag: "1.14.0",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbRedis,
				ImageVersionTag: "7.0.5",
				NodeCount:       6,
			},
			scaleTo: 8,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbKafka,
				ImageVersionTag: "3.1.1",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbKafka,
				ImageVersionTag: "3.2.3",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbRabbitMQ,
				ImageVersionTag: "3.10.9",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbMySQL,
				ImageVersionTag: "8.0.31",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbMongoDB,
				ImageVersionTag: "6.0.3",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbElasticSearch,
				ImageVersionTag: "8.5.2",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbCouchbase,
				ImageVersionTag: "7.1.1",
				NodeCount:       1,
			},
			scaleTo: 2,
		},
	}

	for _, testCase := range testCases {
		tt := testCase
		s.T().Run(fmt.Sprintf("scale-%s-%s-nodes-%v-to-%v", tt.spec.DataServiceName, tt.spec.getImageVersionString(), tt.spec.NodeCount, tt.scaleTo), func(t *testing.T) {
			t.Parallel()

			tt.spec.NamePrefix = fmt.Sprintf("scale-%s-", tt.spec.getImageVersionString())
			deploymentID := s.mustDeployDeploymentSpec(t, tt.spec)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})

			// Create.
			s.mustEnsureDeploymentHealthy(t, deploymentID)
			s.mustEnsureDeploymentInitialized(t, deploymentID)
			s.mustEnsureStatefulSetReady(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)

			// Update.
			updateSpec := tt.spec
			updateSpec.NodeCount = tt.scaleTo
			s.mustUpdateDeployment(t, deploymentID, &updateSpec)
			s.mustEnsureStatefulSetReadyAndUpdatedReplicas(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_ScaleResources() {
	testCases := []struct {
		spec                    ShortDeploymentSpec
		scaleToResourceTemplate string
	}{
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbPostgres,
				ImageVersionTag: "14.6",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbPostgres].ResourceTemplates[1].Name,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbCassandra,
				ImageVersionTag: "4.0.6",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbCassandra].ResourceTemplates[1].Name,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbConsul,
				ImageVersionTag: "1.14.0",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbConsul].ResourceTemplates[1].Name,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbKafka,
				ImageVersionTag: "3.2.3",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbKafka].ResourceTemplates[1].Name,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbRabbitMQ,
				ImageVersionTag: "3.10.9",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbRabbitMQ].ResourceTemplates[1].Name,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbMySQL,
				ImageVersionTag: "8.0.31",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbMySQL].ResourceTemplates[1].Name,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbMongoDB,
				ImageVersionTag: "6.0.2",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbMongoDB].ResourceTemplates[1].Name,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbElasticSearch,
				ImageVersionTag: "8.5.2",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbElasticSearch].ResourceTemplates[1].Name,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbCouchbase,
				ImageVersionTag: "7.1.1",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbCouchbase].ResourceTemplates[1].Name,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbRedis,
				ImageVersionTag: "7.0.5",
				NodeCount:       1,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbRedis].ResourceTemplates[1].Name,
		},
		{
			spec: ShortDeploymentSpec{
				DataServiceName: dbZooKeeper,
				ImageVersionTag: "3.8.0",
				NodeCount:       3,
			},
			scaleToResourceTemplate: s.testPDSTemplatesMap[dbZooKeeper].ResourceTemplates[1].Name,
		},
	}

	for _, testCase := range testCases {
		tt := testCase
		s.T().Run(fmt.Sprintf("scale-%s-%s-resources", tt.spec.DataServiceName, tt.spec.getImageVersionString()), func(t *testing.T) {
			t.Parallel()

			tt.spec.NamePrefix = fmt.Sprintf("scale-%s-", tt.spec.getImageVersionString())
			deploymentID := s.mustDeployDeploymentSpec(t, tt.spec)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})

			// Create.
			s.mustEnsureDeploymentHealthy(t, deploymentID)
			s.mustEnsureDeploymentInitialized(t, deploymentID)
			s.mustEnsureStatefulSetReady(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)

			// Update.
			updateSpec := tt.spec
			updateSpec.ResourceSettingsTemplateName = tt.scaleToResourceTemplate
			s.mustUpdateDeployment(t, deploymentID, &updateSpec)
			s.mustEnsureStatefulSetReadyAndUpdatedReplicas(t, deploymentID)
			s.mustEnsureLoadBalancerServicesReady(t, deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(t, deploymentID)
			s.mustRunBasicSmokeTest(t, deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_Recovery_FromDeletion() {
	deployments := []ShortDeploymentSpec{
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
		s.T().Run(fmt.Sprintf("recover-%s-%s-n%d", deployment.DataServiceName, deployment.getImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("recover-%s-n%d-", deployment.getImageVersionString(), deployment.NodeCount)
			deploymentID := s.mustDeployDeploymentSpec(t, deployment)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})
			s.mustEnsureDeploymentHealthy(t, deploymentID)
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
	deployments := []ShortDeploymentSpec{
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
		s.T().Run(fmt.Sprintf("metrics-%s-%s-n%d", deployment.DataServiceName, deployment.getImageVersionString(), deployment.NodeCount), func(t *testing.T) {
			t.Parallel()

			deployment.NamePrefix = fmt.Sprintf("metrics-%s-n%d-", deployment.getImageVersionString(), deployment.NodeCount)
			deploymentID := s.mustDeployDeploymentSpec(t, deployment)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(t, deploymentID)
				s.waitForDeploymentRemoved(t, deploymentID)
			})
			s.mustEnsureDeploymentHealthy(t, deploymentID)
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
