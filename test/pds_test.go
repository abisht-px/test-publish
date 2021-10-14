package test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.anim.dreamworks.com/DreamCloud/stella-api/api/models"

	client "github.com/portworx/pds-integration-test/test/client"
)

var (
	domainID = uuid.MustParse("95522f98-b216-45e8-a1f5-a0378fffc8bb")
)

type PDSTestSuite struct {
	suite.Suite
	ControlPlane, TargetCluster *client.API
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
	s.ControlPlane = client.NewAPI(controlPlaneAPI)

	targetClusterAPI := mustGetEnvVariable(s.T(), envTargetAPI)
	s.TargetCluster = client.NewAPI(targetClusterAPI)
	s.targetClusterToken = mustGetEnvVariable(s.T(), envTargetToken)
}

func (s *PDSTestSuite) mustReachClusters() {
	s.ControlPlane.MustReachEndpoint(s.T(), endpointTargetClusters)
	// TODO: investigate if we can do a simple sanity check on the target cluster considering AWS authentication restrictions
	// s.TargetCluster.MustReachEndpoint(s.T(), "")
}

func (s *PDSTestSuite) mustRegisterTargetCluster() *models.TargetCluster {
	target := &models.TargetCluster{
		Name:      "target-1",
		APIServer: s.TargetCluster.BaseURL,
		Token:     s.targetClusterToken,
		DomainID:  domainID,
	}
	s.ControlPlane.MustPostJSON(s.T(), endpointTargetClusters, target)
	
	return target
}

func (s* PDSTestSuite) mustCreateEnvironment(targetCluster *models.TargetCluster) *models.Environment {
	environment := &models.Environment{
		Name:             randomName("integration-test-"),
		TargetClusters:   []models.TargetCluster{*targetCluster},
		TargetClusterIDs: []uuid.UUID{targetCluster.ID},
		DomainID:         domainID,
	}
	s.ControlPlane.MustPostJSON(s.T(), endpointEnvironments, environment)

	return environment
}

func (s* PDSTestSuite) mustPopulateDatabaseData(environment *models.Environment) {
	for _, dbDefinition := range DBTypes {
		dbType := dbDefinition.DatabaseType
		dbType.DomainID = domainID
		s.ControlPlane.MustPostJSON(s.T(), endpointDatabaseTypes, dbType)

		for _, versionedImage := range dbDefinition.versionedImages {
			version := versionedImage.Version
			version.DomainID = domainID
			version.DatabaseTypeID = dbType.ID
			version.DatabaseTypeName = dbType.Name
			s.ControlPlane.MustPostJSON(s.T(), endpointVersions, &version)

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
				s.ControlPlane.MustPostJSON(s.T(), endpointImages, image)

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
		s.ControlPlane.MustPostJSON(s.T(), endpointTargetClusters, &targetCluster)
	}
}

func TestPDSSuite(t *testing.T) {
	suite.Run(t, new(PDSTestSuite))
}
