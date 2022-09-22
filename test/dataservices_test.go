package test

import "fmt"

const (
	dbPostgres = "PostgreSQL"
)

func (s *PDSTestSuite) TestPostgreSQL_WriteData() {
	deployments := []ShortDeploymentSpec{
		{
			ServiceName:                  dbPostgres,
			ImageVersionBuild:            "81c330f",
			AppConfigTemplateName:        "QaDefault",
			StorageOptionName:            "QaDefault",
			ResourceSettingsTemplateName: "Qasmall",
			ServiceType:                  "LoadBalancer",
			NamePrefix:                   "autotest-81c330f-",
			NodeCount:                    1,
		},
	}

	for _, deployment := range deployments {
		s.Run(fmt.Sprintf("%s-%s", deployment.ServiceName, deployment.ImageVersionBuild), func() {
			deploymentID := s.mustDeployDeploymentSpec(deployment)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(deploymentID)
				s.mustEnsureDeploymentRemoved(deploymentID)
			})
			s.mustEnsureDeploymentHealthy(deploymentID)
			s.mustEnsureDeploymentInitialized(deploymentID)
			s.mustEnsureStatefulSetReady(deploymentID)
			s.mustReadWriteData(deploymentID)
		})
	}
}

func (s *PDSTestSuite) TestPostgreSQL_UpdateImage() {
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
				NamePrefix:                   "autotest-81c330f-",
				NodeCount:                    1,
			},
			targetVersions: []string{"14.4"},
		},
	}

	for _, tt := range testCases {
		for _, targetVersionTag := range tt.targetVersions {
			s.Run(fmt.Sprintf("update-%s-%s-to-%s", tt.spec.ServiceName, tt.spec.ImageVersionTag, targetVersionTag), func() {
				deploymentID := s.mustDeployDeploymentSpec(tt.spec)
				s.T().Cleanup(func() {
					s.mustRemoveDeployment(deploymentID)
					s.mustEnsureDeploymentRemoved(deploymentID)
				})

				// Create.
				s.mustEnsureDeploymentHealthy(deploymentID)
				s.mustEnsureDeploymentInitialized(deploymentID)
				s.mustEnsureStatefulSetReady(deploymentID)
				s.mustReadWriteData(deploymentID)

				// Update.
				newSpec := tt.spec
				newSpec.ImageVersionTag = targetVersionTag
				s.mustUpdateDeployment(deploymentID, &newSpec)
				s.mustEnsureStatefulSetReady(deploymentID)
				s.mustEnsureStatefulSetImage(deploymentID, targetVersionTag)
				s.mustReadWriteData(deploymentID)
			})
		}
	}
}

func (s *PDSTestSuite) TestPostgreSQL_ScaleUp() {
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
				NamePrefix:                   "autotest-81c330f-",
				NodeCount:                    1,
			},
			scaleTo: 2,
		},
	}

	for _, tt := range testCases {
		s.Run(fmt.Sprintf("scale-%s-%s-nodes-%v-to-%v", tt.spec.ServiceName, tt.spec.ImageVersionTag, tt.spec.NodeCount, tt.scaleTo), func() {
			deploymentID := s.mustDeployDeploymentSpec(tt.spec)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(deploymentID)
				s.mustEnsureDeploymentRemoved(deploymentID)
			})

			// Create.
			s.mustEnsureDeploymentHealthy(deploymentID)
			s.mustEnsureDeploymentInitialized(deploymentID)
			s.mustEnsureStatefulSetReady(deploymentID)
			s.mustReadWriteData(deploymentID)

			// Update.
			updateSpec := tt.spec
			updateSpec.NodeCount = tt.scaleTo
			s.mustUpdateDeployment(deploymentID, &updateSpec)
			s.mustEnsureStatefulSetReady(deploymentID)
			s.mustEnsureStatefulSetReadyReplicas(deploymentID, tt.scaleTo)
			s.mustReadWriteData(deploymentID)
		})
	}
}
