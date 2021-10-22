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
			Name:                 "Cassandra",
			ShortName:            "testcas",
			HasIncrementalBackup: &allowBackups,
			HasFullBackup:        &allowBackups,
		},
		versionedImages: map[string]*versionedImages{
			"3.11.4": {
				images: map[string]*models.Image{
					"af42051": {
						Registry:  imageRegistry,
						Namespace: imageNamespace,
					},
				},
			},
			"3.11.6": {
				images: map[string]*models.Image{
					"af42051": {
						Registry:  imageRegistry,
						Namespace: imageNamespace,
					},
				},
			},
			"3.11.9": {
				images: map[string]*models.Image{
					"af42051": {
						Registry:  imageRegistry,
						Namespace: imageNamespace,
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
	versionedImages map[string]*versionedImages
}

// versionedImages wraps the API Version model together with related images.
// This allows for easier definition of the hierarchy and setting up the reference fields.
// All of the version fields can be derived from context, so the setup takes care of
// initializing the Version field.
type versionedImages struct {
	*models.Version
	images map[string]*models.Image
}
