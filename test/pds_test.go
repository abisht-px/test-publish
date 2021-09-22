package test

import (
	"fmt"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.anim.dreamworks.com/DreamCloud/stella-api/api/models"
)

const (
	domainID = "95522f98-b216-45e8-a1f5-a0378fffc8bb"
)

type PDSTestSuite struct {
	suite.Suite
	controlPlaneAPI, targetClusterAPI, targetClusterToken string
}

func (suite *PDSTestSuite) SetupSuite() {
	// Perform sanity checks.
	suite.mustHaveEnvVariables()
	suite.mustReachClusters()

	// Configure control plane data.
	suite.mustRegisterTargetCluster()
}

func (s *PDSTestSuite) mustHaveEnvVariables() {
	s.controlPlaneAPI = mustGetEnvVariable(s.T(), envControlPlaneAPI)
	s.targetClusterAPI = mustGetEnvVariable(s.T(), envTargetAPI)
	s.targetClusterToken = mustGetEnvVariable(s.T(), envTargetToken)
}

func (s *PDSTestSuite) mustReachClusters() {
	mustReachAddress(s.T(), s.ControlPlaneEndpoint("target-clusters"))
	// TODO: investigate if we can do a simple sanity check on the target cluster considering AWS authentication restrictions
	// mustReachAddress(s.T(), s.targetClusterAPI)
}

func (s *PDSTestSuite) mustRegisterTargetCluster() {
	target := &models.TargetCluster{
		Name:      "target-1",
		APIServer: s.targetClusterAPI,
		Token:     s.targetClusterToken,
		DomainID:  uuid.MustParse(domainID),
	}
	clustersEndpoint := s.ControlPlaneEndpoint("target-clusters")
	mustPostJSON(s.T(), clustersEndpoint, target)
}

func (s *PDSTestSuite) ControlPlaneEndpoint(pathFragments ...string) string {
	return fmt.Sprintf("%s/%s", s.controlPlaneAPI, path.Join(pathFragments...))
}

func TestPDSSuite(t *testing.T) {
	suite.Run(t, new(PDSTestSuite))
}
