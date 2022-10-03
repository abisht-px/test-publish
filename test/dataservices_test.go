package test

import "fmt"

const (
	dbPostgres  = "PostgreSQL"
	dbCassandra = "Cassandra"
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
			NamePrefix:                   "autotest-14.5-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbCassandra,
			ImageVersionTag:              "4.0.5",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "autotest-4.0.5-",
			NodeCount:                    1,
		},
	}

	for _, deployment := range deployments {
		s.Run(fmt.Sprintf("%s-%s", deployment.ServiceName, deployment.getImageVersionString()), func() {
			deploymentID := s.mustDeployDeploymentSpec(deployment)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(deploymentID)
				s.mustEnsureDeploymentRemoved(deploymentID)
			})
			s.mustEnsureDeploymentHealthy(deploymentID)
			s.mustEnsureDeploymentInitialized(deploymentID)
			s.mustEnsureStatefulSetReady(deploymentID)

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
			NamePrefix:                   "autotest-14.5-",
			NodeCount:                    1,
		},
		{
			ServiceName:                  dbCassandra,
			ImageVersionTag:              "4.0.5",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "autotest-4.0.5-",
			NodeCount:                    1,
		},
	}

	for _, deployment := range deployments {
		s.Run(fmt.Sprintf("%s-%s-backup", deployment.ServiceName, deployment.getImageVersionString()), func() {
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
				ServiceName:                  dbCassandra,
				ImageVersionTag:              "4.0.4",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "update-4.0.x-",
				NodeCount:                    1,
			},
			targetVersions: []string{"4.0.5"},
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbPostgres,
				ImageVersionTag:              "14.2",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "autotest-14.2-",
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
				NamePrefix:                   "autotest-14.4-",
				NodeCount:                    1,
			},
			targetVersions: []string{"14.5"},
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
				s.mustRunBasicSmokeTest(deploymentID)

				// Update.
				newSpec := tt.spec
				newSpec.ImageVersionTag = targetVersionTag
				s.mustUpdateDeployment(deploymentID, &newSpec)
				s.mustEnsureStatefulSetImage(deploymentID, targetVersionTag)
				s.mustEnsureStatefulSetReady(deploymentID)
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
				NamePrefix:                   "autotest-14.4-",
				NodeCount:                    1,
			},
			scaleTo: 2,
		},
		{
			spec: ShortDeploymentSpec{
				ServiceName:                  dbCassandra,
				ImageVersionTag:              "4.0.5",
				AppConfigTemplateName:        "QaDefault",
				StorageOptionName:            "QaDefault",
				ResourceSettingsTemplateName: "Qasmall",
				ServiceType:                  "LoadBalancer",
				NamePrefix:                   "autotest-4.0.5-",
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
				NamePrefix:                   "autotest-14.5-",
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
			s.mustRunBasicSmokeTest(deploymentID)

			// Update.
			updateSpec := tt.spec
			updateSpec.NodeCount = tt.scaleTo
			s.mustUpdateDeployment(deploymentID, &updateSpec)
			s.mustEnsureStatefulSetReady(deploymentID)
			s.mustEnsureStatefulSetReadyReplicas(deploymentID, tt.scaleTo)
			s.mustRunBasicSmokeTest(deploymentID)
		})
	}
}
