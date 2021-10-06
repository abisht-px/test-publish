package test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.anim.dreamworks.com/DreamCloud/stella-api/api/models"
)

var (
	domainID = uuid.MustParse("95522f98-b216-45e8-a1f5-a0378fffc8bb")
)

type PDSTestSuite struct {
	suite.Suite
	ControlPlane, TargetCluster api
	targetClusterToken string
}

func (suite *PDSTestSuite) SetupSuite() {
	// Perform sanity checks.
	suite.mustHaveEnvVariables()
	suite.mustReachClusters()

	// Configure control plane data.
	targetCluster := suite.mustRegisterTargetCluster()
	testEnvironment := suite.mustCreateEnvironment(targetCluster)
	suite.mustPopulateDatabaseData(testEnvironment)
}

func (s *PDSTestSuite) mustHaveEnvVariables() {
	controlPlaneAPI := mustGetEnvVariable(s.T(), envControlPlaneAPI)
	s.ControlPlane = api{baseURL: controlPlaneAPI}

	targetClusterAPI := mustGetEnvVariable(s.T(), envTargetAPI)
	s.TargetCluster = api{baseURL: targetClusterAPI}
	s.targetClusterToken = mustGetEnvVariable(s.T(), envTargetToken)
}

func (s *PDSTestSuite) mustReachClusters() {
	mustReachAddress(s.T(), s.ControlPlane.endpoint("target-clusters"))
	// TODO: investigate if we can do a simple sanity check on the target cluster considering AWS authentication restrictions
	// mustReachAddress(s.T(), s.targetClusterAPI)
}

func (s *PDSTestSuite) mustRegisterTargetCluster() *models.TargetCluster {
	target := &models.TargetCluster{
		Name:      "target-1",
		APIServer: s.TargetCluster.baseURL,
		Token:     s.targetClusterToken,
		DomainID:  domainID,
	}
	clustersEndpoint := s.ControlPlane.endpoint("target-clusters")
	mustPostJSON(s.T(), clustersEndpoint, target)
	
	return target
}

func (s* PDSTestSuite) mustCreateEnvironment(targetCluster *models.TargetCluster) *models.Environment {
	environment := &models.Environment{
		Name:             randomName("integration-test-"),
		TargetClusters:   []models.TargetCluster{*targetCluster},
		TargetClusterIDs: []uuid.UUID{targetCluster.ID},
		DomainID:         domainID,
	}
	envEndpoint := s.ControlPlane.endpoint("environments")
	mustPostJSON(s.T(), envEndpoint, environment)

	return environment
}

func (s* PDSTestSuite) mustPopulateDatabaseData(environment *models.Environment) {
	dbTypesEndpoint := s.ControlPlane.endpoint("types")
	dbVersionsEndpoint := s.ControlPlane.endpoint("versions")
	imagesEndpoint := s.ControlPlane.endpoint("images")
	targetClustersEndpoint := s.ControlPlane.endpoint("target-clusters")

	for _, dbDefinition := range DBTypes {
		dbType := dbDefinition.DatabaseType
		dbType.DomainID = domainID
		mustPostJSON(s.T(), dbTypesEndpoint, dbType)

		for _, versionedImage := range dbDefinition.versionedImages {
			version := versionedImage.Version
			version.DomainID = domainID
			version.DatabaseTypeID = dbType.ID
			version.DatabaseTypeName = dbType.Name
			mustPostJSON(s.T(), dbVersionsEndpoint, &version)

			for _, image := range versionedImage.images {
				image.DomainID = domainID
				image.Name = "pds-" + dbType.Name
				image.Tag = version.Name
				image.Environments = environment.Name
				image.DatabaseType = dbType
				image.DatabaseTypeID = dbType.ID
				image.DatabaseTypeName = dbType.Name
				image.Version = version
				image.VersionID = version.ID
				image.VersionName = version.Name
				mustPostJSON(s.T(), imagesEndpoint, image)

				// Add image to every target cluster of the test environment.
				for _, targetCluster := range environment.TargetClusters {
					targetCluster.Images = append(targetCluster.Images, *image)
					targetCluster.ImageIDs = append(targetCluster.ImageIDs, image.ID)
				}
			}
		}
	}

	// Save all target clusters with attached images.
	for _, targetCluster := range environment.TargetClusters {
		mustPostJSON(s.T(), targetClustersEndpoint, &targetCluster)
	}
}

func TestPDSSuite(t *testing.T) {
	suite.Run(t, new(PDSTestSuite))
}
