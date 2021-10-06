package seeding

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

var (
	allowBackups = true
	dbTypes      = []DatabaseType{
		{
			Name:                 "Cassandra",
			ShortName:            "cas",
			HasIncrementalBackup: &allowBackups,
			HasFullBackup:        &allowBackups,
			Templates: []Template{{
				Name:      "Small",
				SortOrder: 0,
				Resources: map[string]interface{}{
					"cpu":    "1",
					"memory": "2Gi",
					"disk":   "1Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "authorizer",
						DefaultValue: "AllowAllAuthorizer",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "authenticator",
						DefaultValue: "AllowAllAuthenticator",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "maxHeapSize",
						DefaultValue: "1G",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "heapNewSize",
						DefaultValue: "400M",
						Type:         "string",
						Required:     true,
					},
				},
			}, {
				Name:      "Medium",
				SortOrder: 1,
				Resources: map[string]interface{}{
					"cpu":    "1",
					"memory": "3Gi",
					"disk":   "2Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "authorizer",
						DefaultValue: "AllowAllAuthorizer",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "authenticator",
						DefaultValue: "AllowAllAuthenticator",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "maxHeapSize",
						DefaultValue: "2G",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "heapNewSize",
						DefaultValue: "400M",
						Type:         "string",
						Required:     true,
					},
				},
			}, {
				Name:      "Large",
				SortOrder: 2,
				Resources: map[string]interface{}{
					"cpu":    "2",
					"memory": "4Gi",
					"disk":   "3Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "authorizer",
						DefaultValue: "AllowAllAuthorizer",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "authenticator",
						DefaultValue: "AllowAllAuthenticator",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "maxHeapSize",
						DefaultValue: "3G",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "heapNewSize",
						DefaultValue: "400M",
						Type:         "string",
						Required:     true,
					},
				},
			}},
			Versions: []Version{
				{
					Name: "3.11.4",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-cassandra",
							Tag:          "3.11.4",
							Build:        "af42051",
							Environments: "develop",
						},
					},
				},
				{
					Name: "3.11.6",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-cassandra",
							Tag:          "3.11.6",
							Build:        "af42051",
							Environments: "develop",
						},
					},
				},
				{
					Name: "3.11.9",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-cassandra",
							Tag:          "3.11.9",
							Build:        "af42051",
							Environments: "develop",
						},
					},
				},
			},
			ComingSoon: false,
		},
		{
			Name:       "Elasticsearch",
			ShortName:  "ess",
			ComingSoon: true,
		},
		{
			Name:       "Couchbase",
			ShortName:  "cbs",
			ComingSoon: true,
		},
		{
			Name:       "Consul",
			ShortName:  "con",
			ComingSoon: true,
		},
		{
			Name:       "DatastaxEnterprise",
			ShortName:  "dse",
			ComingSoon: true,
		},
		{
			Name:      "Kafka",
			ShortName: "kf",
			Templates: []Template{{
				Name:      "Small",
				SortOrder: 0,
				Resources: map[string]interface{}{
					"cpu":    "0.5",
					"memory": "1Gi",
					"disk":   "1Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "ZOOKEEPER_CONNECTION_STRING",
						DefaultValue: "",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "heapSize",
						DefaultValue: "400M",
						Type:         "string",
						Required:     true,
					},
				},
			}, {
				Name:      "Medium",
				SortOrder: 1,
				Resources: map[string]interface{}{
					"cpu":    "1",
					"memory": "2Gi",
					"disk":   "2Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "ZOOKEEPER_CONNECTION_STRING",
						DefaultValue: "",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "heapSize",
						DefaultValue: "600M",
						Type:         "string",
						Required:     true,
					},
				},
			}, {
				Name:      "Large",
				SortOrder: 2,
				Resources: map[string]interface{}{
					"cpu":    "2",
					"memory": "3Gi",
					"disk":   "3Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "ZOOKEEPER_CONNECTION_STRING",
						DefaultValue: "",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "heapSize",
						DefaultValue: "800M",
						Type:         "string",
						Required:     true,
					},
				},
			}},
			Versions: []Version{
				{
					Name: "2.4.1",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-kafka",
							Tag:          "2.4.1",
							Build:        "d4a60bc",
							Environments: "develop",
						},
					},
				},
				{
					Name: "2.7.0",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-kafka",
							Tag:          "2.7.0",
							Build:        "d4a60bc",
							Environments: "develop",
						},
					},
				},
			},
			ComingSoon: false,
		},
		{
			Name:       "Mongodb",
			ShortName:  "mdb",
			ComingSoon: true,
		},
		{
			Name:      "Rabbitmq",
			ShortName: "rmq",
			Templates: []Template{{
				Name:      "Small",
				SortOrder: 0,
				Resources: map[string]interface{}{
					"cpu":    "0.5",
					"memory": "1Gi",
					"disk":   "1Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "datadogPass",
						DefaultValue: "datadog",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "defaultPass",
						DefaultValue: "defaultpass",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "defaultUser",
						DefaultValue: "defaultuser",
						Type:         "string",
						Required:     true,
					},
				},
			}, {
				Name:      "Medium",
				SortOrder: 1,
				Resources: map[string]interface{}{
					"cpu":    "1",
					"memory": "2Gi",
					"disk":   "2Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "datadogPass",
						DefaultValue: "datadog",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "defaultPass",
						DefaultValue: "defaultpass",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "defaultUser",
						DefaultValue: "defaultuser",
						Type:         "string",
						Required:     true,
					},
				},
			}, {
				Name:      "Large",
				SortOrder: 2,
				Resources: map[string]interface{}{
					"cpu":    "2",
					"memory": "3Gi",
					"disk":   "3Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "datadogPass",
						DefaultValue: "datadog",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "defaultPass",
						DefaultValue: "defaultpass",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "defaultUser",
						DefaultValue: "defaultuser",
						Type:         "string",
						Required:     true,
					},
				},
			}},
			Versions: []Version{
				{
					Name: "3.8.9",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-rabbitmq",
							Tag:          "3.8.9",
							Build:        "11d996b",
							Environments: "develop",
						},
					},
				},
			},
			ComingSoon: false,
		},
		{
			Name:          "Redis",
			ShortName:     "red",
			HasFullBackup: &allowBackups,
			Templates: []Template{{
				Name:      "Small",
				SortOrder: 0,
				Resources: map[string]interface{}{
					"cpu":    "0.5",
					"memory": "1Gi",
					"disk":   "1Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "REDIS_DIR",
						DefaultValue: "/data",
						Type:         "string",
						Required:     true,
					},
				},
			}, {
				Name:      "Medium",
				SortOrder: 1,
				Resources: map[string]interface{}{
					"cpu":    "1",
					"memory": "2Gi",
					"disk":   "2Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "REDIS_DIR",
						DefaultValue: "/data",
						Type:         "string",
						Required:     true,
					},
				},
			}, {
				Name:      "Large",
				SortOrder: 2,
				Resources: map[string]interface{}{
					"cpu":    "2",
					"memory": "3Gi",
					"disk":   "3Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "REDIS_DIR",
						DefaultValue: "/data",
						Type:         "string",
						Required:     true,
					},
				},
			}},
			Versions: []Version{
				{
					Name: "3.2.6",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-redis",
							Tag:          "3.2.6",
							Build:        "206c119",
							Environments: "develop",
						},
					},
				},
				{
					Name: "3.2.7",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-redis",
							Tag:          "3.2.7",
							Build:        "206c119",
							Environments: "develop",
						},
					},
				},
				{
					Name: "5.0.4",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-redis",
							Tag:          "5.0.4",
							Build:        "206c119",
							Environments: "develop",
						},
					},
				},
				{
					Name: "5.0.9",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-redis",
							Tag:          "5.0.9",
							Build:        "206c119",
							Environments: "develop",
						},
					},
				},
			},
			ComingSoon: false,
		},
		{
			Name:      "Zookeeper",
			ShortName: "zk",
			Templates: []Template{{
				Name:      "Small",
				SortOrder: 0,
				Resources: map[string]interface{}{
					"cpu":    "0.5",
					"memory": "1Gi",
					"disk":   "1Gi",
				},
			}, {
				Name:      "Medium",
				SortOrder: 1,
				Resources: map[string]interface{}{
					"cpu":    "1",
					"memory": "2Gi",
					"disk":   "2Gi",
				},
			}, {
				Name:      "Large",
				SortOrder: 2,
				Resources: map[string]interface{}{
					"cpu":    "2",
					"memory": "3Gi",
					"disk":   "3Gi",
				},
			}},
			Versions: []Version{
				{
					Name: "3.4.14",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-zookeeper",
							Tag:          "3.4.14",
							Build:        "dc61312",
							Environments: "develop",
						},
					},
				},
			},
			ComingSoon: false,
		},
	}
)

var (
	// defaultDomainID is a hardcoded value that matches the dashboard's domain_id
	// https://github.com/portworx/DreamCloud_dc-dashboard/blob/d8faf8de3b041974a9d27b1a22bce7702cc82ad8/src/auth/index.js#L61
	defaultDomainID = uuid.MustParse("95522f98-b216-45e8-a1f5-a0378fffc8bb")
)

func InsertDefaultDbTypes(db *gorm.DB) error {
	opts := &gormigrate.Options{
		UseTransaction: true,
	}
	m := gormigrate.New(db, opts, []*gormigrate.Migration{
		{
			ID: "202108170000",
			Migrate: func(gdb *gorm.DB) error {
				// initialize DB types
				for _, databaseType := range dbTypes {
					databaseType.DomainID = defaultDomainID
					if err := db.Create(&databaseType).Error; err != nil {
						return err
					}

					// initialize all versions of DB type
					for _, version := range databaseType.Versions {
						version.DomainID = defaultDomainID
						version.DatabaseTypeID = databaseType.ID
						version.DatabaseTypeName = databaseType.Name
						if err := db.Create(&version).Error; err != nil {
							return err
						}

						// initialize all images of version
						for _, image := range version.Images {
							image.DomainID = defaultDomainID
							image.DatabaseTypeID = databaseType.ID
							image.DatabaseTypeName = databaseType.Name
							image.VersionID = version.ID
							image.VersionName = version.Name
							if err := db.Create(&image).Error; err != nil {
								return err
							}
						}
					}

					// initialize templates of DB type
					for _, template := range databaseType.Templates {
						template.DomainID = defaultDomainID
						template.DatabaseTypeID = databaseType.ID
						template.DatabaseTypeName = databaseType.Name
						if err := db.Create(&template).Error; err != nil {
							return err
						}
					}
				}
				return nil
			},
			Rollback: func(gdb *gorm.DB) error {
				// skip rollback for now
				return nil
			},
		},
	})
	return m.Migrate()
}
