package test

import (
	"github.anim.dreamworks.com/DreamCloud/stella-api/api/models"
)

const (
	imageRegistry  = "docker.io"
	imageNamespace = "portworx"
)

var (
	allowBackups = true

	DBTypes = []*dbDefinition{
		Cassandra,
	}
	Cassandra = &dbDefinition{
		DatabaseType: &models.DatabaseType{
			Name:                 "TestCassandra",
			ShortName:            "testcas",
			HasIncrementalBackup: &allowBackups,
			HasFullBackup:        &allowBackups,
		},
		versionedImages: []*versionedImages{
			{
				Version: &models.Version{
					Name: "3.11.4",
				},
				images: []*models.Image{
					{
						Registry:  imageRegistry,
						Namespace: imageNamespace,
						Build:     "af42051",
					},
				},
			},
			{
				Version: &models.Version{
					Name: "3.11.6",
				},
				images: []*models.Image{
					{
						Registry:  imageRegistry,
						Namespace: imageNamespace,
						Build:     "af42051",
					},
				},
			},
			{
				Version: &models.Version{
					Name: "3.11.9",
				},
				images: []*models.Image{
					{
						Registry:  imageRegistry,
						Namespace: imageNamespace,
						Build:     "af42051",
					},
				},
			},
		},
	}
)

// dbDefinition wraps the API DatabaseType and Version models together with related images.
// This allows for easier definition of the hierarchy and setting up the reference fields.
type dbDefinition struct {
	*models.DatabaseType
	versionedImages []*versionedImages
}

// versionedImages wraps the API Version model together with related images.
// This allows for easier definition of the hierarchy and setting up the reference fields.
type versionedImages struct {
	*models.Version
	images []*models.Image
}
