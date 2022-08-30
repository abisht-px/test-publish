package test

import (
	"fmt"

	_ "github.com/stretchr/testify/suite"
)

const (
	DeploymentSpecIDExample01 PDSDeploymentSpecID = iota
)

// TODO: Remove this test after the first real test is added.
func (s *PDSTestSuite) TestDeploymentCreateDelete() {
	tests := []struct {
		specID PDSDeploymentSpecID
	}{
		{
			specID: DeploymentSpecIDExample01,
		},
	}

	for _, tt := range tests {
		s.Run(fmt.Sprintf("Spec %d", tt.specID), func() {
			s.Require().Contains(s.shortDeploymentSpecMap, tt.specID, "Deployment spec %s not found.", tt.specID)
			deployment := s.shortDeploymentSpecMap[tt.specID]

			deploymentID := s.mustDeployDeploymentSpec(deployment)
			s.T().Cleanup(func() {
				s.mustRemoveDeployment(deploymentID)
				s.mustEnsureDeploymentRemoved(deploymentID)
			})
			s.mustEnsureDeploymentHealty(deploymentID)
			s.T().Logf("Good deployment %s for spec %d", deploymentID, tt.specID)
		})
	}
}
