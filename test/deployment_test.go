package test

import (
	"time"

	"github.anim.dreamworks.com/DreamCloud/stella-api/api/models"
	"github.com/google/uuid"
)

func (s *PDSTestSuite) TestCassandraDeploy() {
	// Given:
	dbDefinition := Cassandra
	versionedImage := dbDefinition.versionedImages["3.11.9"]
	imageBuild := versionedImage.images["af42051"]
	randomName := randomName("cas-test-")
	deployment := &models.Deployment{
		ClusterName:     randomName,
		NodeCount:       3,
		Service:         randomName + "-service",
		StorageProvider: "portworx",
		Configuration: map[string]interface{}{
			"maxHeapSize": "2G",
			"heapNewSize": "400M",
		},
		Resources: map[string]interface{}{
			"cpu":     "1",
			"memory":  "4G",
			"storage": "1G",
			"disk":    "1G",
		},
		DomainID:        domainID,
		ProjectID:       uuid.New(),
		DatabaseTypeID:  dbDefinition.DatabaseType.ID,
		VersionID:       versionedImage.Version.ID,
		ImageID:         &imageBuild.ID,
		EnvironmentID:   s.TestEnvironment.ID,
		TargetClusterID: &s.TargetCluster.Model.ID,
	}

	// When:
	s.ControlPlane.API.MustPostJSON(s.T(), "deployments", deployment)

	// Then:
	s.Assert().Eventuallyf(
		func() bool {
			// TODO (DS-485): Add a real check for Cassandra health.
			return false
		},
		// TODO (DS-485): extend timeout to a meaningful timespan once there's a real check
		0*time.Minute,
		5*time.Second,
		"Waiting for deployment:\nDB Type: %s\nVersion: %s\nImage: %s\nEnvironment: %s",
		dbDefinition.Name, versionedImage.Version.Name, imageBuild.Name, s.TestEnvironment.Name)
}
