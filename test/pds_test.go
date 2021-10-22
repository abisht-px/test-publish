package test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.anim.dreamworks.com/DreamCloud/stella-api/api/models"

	client "github.com/portworx/pds-integration-test/test/client"
	cluster "github.com/portworx/pds-integration-test/test/cluster"
	"github.com/portworx/pds-integration-test/test/color"
)

var (
	domainID = uuid.New()
)

type PDSTestSuite struct {
	suite.Suite
	ControlPlane    *cluster.ControlPlane
	TargetCluster   *cluster.Target
	TestEnvironment *models.Environment
}

func (s *PDSTestSuite) SetupSuite() {
	// Perform basic setup with sanity checks.
	env := mustHaveEnvVariables(s.T())
	s.mustHaveControlPlane(env)
	s.mustHaveTargetCluster(env)
	s.mustReachClusters()

	// Configure control plane data.
	targetCluster := s.mustRegisterTargetCluster(s.TargetCluster)
	testEnvironment := s.mustCreateEnvironment(targetCluster)
	s.mustPopulateDatabaseData(testEnvironment)
}

func (s *PDSTestSuite) AfterTest(suiteName, testName string) {
	if s.T().Failed() {
		s.T().Log(color.Red(fmt.Sprintf("Failed test %s:", testName)))
		s.ControlPlane.LogStatus(s.T())
		s.TargetCluster.LogStatus(s.T(), s.TestEnvironment.Name)
	}
}

func (s *PDSTestSuite) mustHaveControlPlane(env environment) {
	apiClient := client.NewAPI(env.controlPlaneAPI)
	kubeContext := env.controlPlaneKubeconfig
	s.ControlPlane = cluster.NewControlPlane(apiClient, kubeContext)
}

func (s *PDSTestSuite) mustHaveTargetCluster(env environment) {
	apiClient := client.NewAPI(env.targetAPI)

	config, err := clientcmd.BuildConfigFromFlags("", env.targetKubeconfig)
	s.Require().NoErrorf(err, "Loading kubeconfig from path %q.", env.targetKubeconfig)
	clientset := kubernetes.NewForConfigOrDie(config)

	s.TargetCluster = cluster.NewTarget(env.targetToken, env.targetKubeconfig, apiClient, clientset)
}

func (s *PDSTestSuite) mustReachClusters() {
	s.ControlPlane.API.MustReachEndpoint(s.T(), endpointTargetClusters)
	s.T().Log(color.Cyan("Target nodes:"))
	s.T().Log(s.TargetCluster.MustKubectl(s.T(), "get", "nodes"))
	// TODO: investigate if we can do a simple sanity check on the target cluster API considering AWS authentication restrictions.
	// s.TargetCluster.MustReachEndpoint(s.T(), "")
}

func (s *PDSTestSuite) mustRegisterTargetCluster(tc *cluster.Target) *models.TargetCluster {
	target := &models.TargetCluster{
		Name:      "target-1",
		APIServer: tc.API.BaseURL,
		Token:     tc.Token,
		DomainID:  domainID,
	}
	s.ControlPlane.API.MustPostJSON(s.T(), endpointTargetClusters, target)
	s.TargetCluster.Model = target

	return target
}

func (s *PDSTestSuite) mustCreateEnvironment(targetCluster *models.TargetCluster) *models.Environment {
	environment := &models.Environment{
		Name:             randomName("integration-test-"),
		TargetClusters:   []models.TargetCluster{*targetCluster},
		TargetClusterIDs: []uuid.UUID{targetCluster.ID},
		DomainID:         domainID,
	}
	s.ControlPlane.API.MustPostJSON(s.T(), endpointEnvironments, environment)
	s.TestEnvironment = environment

	return environment
}

func (s *PDSTestSuite) mustPopulateDatabaseData(environment *models.Environment) {
	for _, dbDefinition := range DBTypes {
		dbType := dbDefinition.DatabaseType
		dbType.DomainID = domainID
		s.ControlPlane.API.MustPostJSON(s.T(), endpointDatabaseTypes, dbType)

		for versionName, versionedImage := range dbDefinition.versionedImages {
			version := &models.Version{
				DomainID:         domainID,
				Name:             versionName,
				DatabaseTypeID:   dbType.ID,
				DatabaseTypeName: dbType.Name,
			}
			versionedImage.Version = version
			s.ControlPlane.API.MustPostJSON(s.T(), endpointVersions, version)

			for build, image := range versionedImage.images {
				image.DomainID = domainID
				image.Name = "pds-" + dbType.Name
				image.Build = build
				image.Tag = version.Name
				image.Environments = environment.Name
				image.DatabaseType = dbType
				image.DatabaseTypeID = dbType.ID
				image.DatabaseTypeName = dbType.Name
				image.Version = version
				image.VersionID = version.ID
				image.VersionName = version.Name
				s.ControlPlane.API.MustPostJSON(s.T(), endpointImages, image)

				// Add image to registered images of test target cluster.
				// Also append imageID since they're not automatically synchronized by the server.
				s.TargetCluster.Model.Images = append(s.TargetCluster.Model.Images, *image)
				s.TargetCluster.Model.ImageIDs = append(s.TargetCluster.Model.ImageIDs, image.ID)
			}
		}
	}
	s.ControlPlane.API.MustPostJSON(s.T(), endpointTargetClusters, s.TargetCluster.Model)
}

func TestPDSSuite(t *testing.T) {
	suite.Run(t, new(PDSTestSuite))
}
