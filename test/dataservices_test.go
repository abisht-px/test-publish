package test

import (
	"fmt"
	"time"
)

const (
	dbPostgres  = "PostgreSQL"
	dbCassandra = "Cassandra"
	dbRedis     = "Redis"
	dbKafka     = "Kafka"
	dbRabbitMQ  = "RabbitMQ"
	dbZooKeeper = "ZooKeeper"
)

func (s *PDSTestSuite) TestDataService_WriteData() {
	deployments := []ShortDeploymentSpec{
		{
			ServiceName:                  dbPostgres,
			ImageVersionTag:              "14.5",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-14.5-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbCassandra,
			ImageVersionTag:              "4.0.6",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-4.0.6-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbRedis,
			ImageVersionTag:              "7.0.5",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-7.0.5-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbZooKeeper,
			ImageVersionTag:              "3.6.3",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.6.3-n3-",
			NodeCount:                    3,
		},
		{
			ServiceName:                  dbZooKeeper,
			ImageVersionTag:              "3.7.1",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.7.1-n3-",
			NodeCount:                    3,
		},
		{
			ServiceName:                  dbZooKeeper,
			ImageVersionTag:              "3.8.0",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.8.0-n3-",
			NodeCount:                    3,
		},
		{
			ServiceName:                  dbKafka,
			ImageVersionTag:              "3.0.1",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.0.1-n1-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbKafka,
			ImageVersionTag:              "3.0.1",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.0.1-n3-",
			NodeCount:                    3,
		},
		{
			ServiceName:                  dbKafka,
			ImageVersionTag:              "3.1.1",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.1.1-n1-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbKafka,
			ImageVersionTag:              "3.1.1",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.1.1-n3-",
			NodeCount:                    3,
		},
		{
			ServiceName:                  dbKafka,
			ImageVersionTag:              "3.2.1",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.2.1-n1-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbKafka,
			ImageVersionTag:              "3.2.1",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.2.1-n3-",
			NodeCount:                    3,
		},
		{
			ServiceName:                  dbRabbitMQ,
			ImageVersionTag:              "3.9.22",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.9.22-n1-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbRabbitMQ,
			ImageVersionTag:              "3.10.6",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.10.6-n1-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbRabbitMQ,
			ImageVersionTag:              "3.10.7",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.10.7-n1-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbRabbitMQ,
			ImageVersionTag:              "3.10.9",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "write-3.10.9-n1-",
			NodeCount:                    1,
		},
	}

	for _, deployment := range deployments {
		s.Run(fmt.Sprintf("write-%s-%s-n%d", deployment.ServiceName, deployment.getImageVersionString(), deployment.NodeCount), func() {
			deploymentID := s.mustDeployDeploymentSpec(deployment)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(deploymentID)
				s.mustEnsureDeploymentRemoved(deploymentID)
			})
			s.mustEnsureDeploymentHealthy(deploymentID)
			s.mustEnsureDeploymentInitialized(deploymentID)
			s.mustEnsureStatefulSetReady(deploymentID)
			s.mustEnsureLoadBalancerServicesReady(deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(deploymentID)

			s.mustRunBasicSmokeTest(deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestDataService_Backup() {
	deployments := []ShortDeploymentSpec{
		{
			ServiceName:                  dbPostgres,
			ImageVersionTag:              "14.5",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "backup-14.5-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbCassandra,
			ImageVersionTag:              "4.0.6",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "backup-4.0.6-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbRedis,
			ImageVersionTag:              "7.0.5",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "backup-7.0.5-",
			NodeCount:                    1,
		},
	}

	for _, deployment := range deployments {
		s.Run(fmt.Sprintf("backup-%s-%s", deployment.ServiceName, deployment.getImageVersionString()), func() {
			deploymentID := s.mustDeployDeploymentSpec(deployment)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(deploymentID)
				s.mustEnsureDeploymentRemoved(deploymentID)
			})
			s.mustEnsureDeploymentHealthy(deploymentID)
			s.mustEnsureDeploymentInitialized(deploymentID)
			s.mustEnsureStatefulSetReady(deploymentID)

			backup := s.mustCreateBackup(deploymentID)
			s.mustEnsureBackupSuccessful(deploymentID, backup.GetClusterResourceName())
			s.mustDeleteBackup(backup.GetId())
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
				ServiceName:                  dbPostgres,
				ImageVersionTag:              "14.2",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-14.2-",
				NodeCount:                    1,
			},
			targetVersions: []string{"14.4", "14.5"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbPostgres,
				ImageVersionTag:              "14.4",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-14.4-",
				NodeCount:                    1,
			},
			targetVersions: []string{"14.5"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbCassandra,
				ImageVersionTag:              "4.0.5",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-4.0.x-",
				NodeCount:                    1,
			},
			targetVersions: []string{"4.0.6"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRedis,
				ImageVersionTag:              "7.0.0",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-7.0.0-",
				NodeCount:                    1,
			},
			targetVersions: []string{"7.0.2", "7.0.4", "7.0.5"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRedis,
				ImageVersionTag:              "7.0.2",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-7.0.2-",
				NodeCount:                    1,
			},
			targetVersions: []string{"7.0.4", "7.0.5"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRedis,
				ImageVersionTag:              "7.0.4",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-7.0.4-",
				NodeCount:                    1,
			},
			targetVersions: []string{"7.0.5"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbKafka,
				ImageVersionTag:              "3.2.0",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-3.2.0-",
				NodeCount:                    1,
			},
			targetVersions: []string{"3.2.3"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbKafka,
				ImageVersionTag:              "3.2.1",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-3.2.1-",
				NodeCount:                    1,
			},
			targetVersions: []string{"3.2.3"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRabbitMQ,
				ImageVersionTag:              "3.9.21",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-3.9.21-",
				NodeCount:                    1,
			},
			targetVersions: []string{"3.9.22"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRabbitMQ,
				ImageVersionTag:              "3.10.6",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-3.10.6-",
				NodeCount:                    1,
			},
			targetVersions: []string{"3.10.9"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRabbitMQ,
				ImageVersionTag:              "3.10.7",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-3.10.7-",
				NodeCount:                    1,
			},
			targetVersions: []string{"3.10.9"},
		},
	}

	for _, tt := range testCases {
		for _, targetVersionTag := range tt.targetVersions {
			s.Run(fmt.Sprintf("update-%s-%s-to-%s", tt.spec.ServiceName, tt.spec.getImageVersionString(), targetVersionTag), func() {
				deploymentID := s.mustDeployDeploymentSpec(tt.spec)
				s.T().Cleanup(func() {
					s.mustRemoveDeployment(deploymentID)
					s.mustEnsureDeploymentRemoved(deploymentID)
				})

				// Create.
				s.mustEnsureDeploymentHealthy(deploymentID)
				s.mustEnsureDeploymentInitialized(deploymentID)
				s.mustEnsureStatefulSetReady(deploymentID)
				s.mustEnsureLoadBalancerServicesReady(deploymentID)
				s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(deploymentID)
				s.mustRunBasicSmokeTest(deploymentID)

				// Update.
				newSpec := tt.spec
				newSpec.ImageVersionTag = targetVersionTag
				s.mustUpdateDeployment(deploymentID, &newSpec)
				s.mustEnsureStatefulSetImage(deploymentID, targetVersionTag)
				s.mustEnsureStatefulSetReady(deploymentID)
				s.mustEnsureLoadBalancerServicesReady(deploymentID)
				s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(deploymentID)
				time.Sleep(10 * time.Second)
				s.mustRunBasicSmokeTest(deploymentID)
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
				ServiceName:                  dbPostgres,
				ImageVersionTag:              "14.4",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scaleup-14.4-",
				NodeCount:                    1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbPostgres,
				ImageVersionTag:              "14.5",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scaleup-14.5-",
				NodeCount:                    1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbCassandra,
				ImageVersionTag:              "4.0.6",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scaleup-4.0.6-",
				NodeCount:                    1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRedis,
				ImageVersionTag:              "7.0.5",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scaleup-7.0.5-",
				NodeCount:                    6,
			},
			scaleTo: 8,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbKafka,
				ImageVersionTag:              "3.0.1",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scaleup-3.0.1-",
				NodeCount:                    1,
			},
			scaleTo: 3,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbKafka,
				ImageVersionTag:              "3.1.1",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scaleup-3.1.1-",
				NodeCount:                    1,
			},
			scaleTo: 3,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbKafka,
				ImageVersionTag:              "3.2.1",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scaleup-3.2.1-",
				NodeCount:                    1,
			},
			scaleTo: 3,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRabbitMQ,
				ImageVersionTag:              "3.9.22",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scale-3.9.22-",
				NodeCount:                    1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRabbitMQ,
				ImageVersionTag:              "3.10.6",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scale-3.10.6-",
				NodeCount:                    1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRabbitMQ,
				ImageVersionTag:              "3.10.7",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scale-3.10.7-",
				NodeCount:                    1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbRabbitMQ,
				ImageVersionTag:              "3.10.9",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "scale-3.10.9-",
				NodeCount:                    1,
			},
			scaleTo: 2,
		},
	}

	for _, tt := range testCases {
		s.Run(fmt.Sprintf("scale-%s-%s-nodes-%v-to-%v", tt.spec.ServiceName, tt.spec.getImageVersionString(), tt.spec.NodeCount, tt.scaleTo), func() {
			deploymentID := s.mustDeployDeploymentSpec(tt.spec)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(deploymentID)
				s.mustEnsureDeploymentRemoved(deploymentID)
			})

			// Create.
			s.mustEnsureDeploymentHealthy(deploymentID)
			s.mustEnsureDeploymentInitialized(deploymentID)
			s.mustEnsureStatefulSetReady(deploymentID)
			s.mustEnsureLoadBalancerServicesReady(deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(deploymentID)
			s.mustRunBasicSmokeTest(deploymentID)

			// Update.
			updateSpec := tt.spec
			updateSpec.NodeCount = tt.scaleTo
			s.mustUpdateDeployment(deploymentID, &updateSpec)
			s.mustEnsureStatefulSetReadyAndUpdatedReplicas(deploymentID, tt.scaleTo)
			s.mustEnsureLoadBalancerServicesReady(deploymentID)
			s.mustEnsureLoadBalancerHostsAccessibleIfNeeded(deploymentID)
			s.mustRunBasicSmokeTest(deploymentID)
		})
	}
}
