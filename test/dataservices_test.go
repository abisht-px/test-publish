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
			s.mustEnsureDeploymentReady(deploymentID)
			s.mustReadWriteData(deploymentID)
		})
	}
}
