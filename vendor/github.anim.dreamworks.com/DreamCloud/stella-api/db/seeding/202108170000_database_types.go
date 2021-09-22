package seeding

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
)

var (
	dbTypes = []DatabaseType{
		{
			Name:      "Cassandra",
			ShortName: "cas",
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
		},
		{
			Name:      "Elasticsearch",
			ShortName: "ess",
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
						Name:         "heapSize",
						DefaultValue: "1G",
						Type:         "string",
						Required:     true,
					},
				},
			}},
			Versions: []Version{
				{
					Name: "5.4.3",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-elasticsearch",
							Tag:          "5.4.3",
							Build:        "5d1d0e6",
							Environments: "develop",
						},
					},
				},
				{
					Name: "6.3.2",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-elasticsearch",
							Tag:          "6.3.2",
							Build:        "5d1d0e6",
							Environments: "develop",
						},
					},
				},
				{
					Name: "6.7.1",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-elasticsearch",
							Tag:          "6.7.1",
							Build:        "5d1d0e6",
							Environments: "develop",
						},
					},
				},
				{
					Name: "7.12.1",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-elasticsearch",
							Tag:          "7.12.1",
							Build:        "5d1d0e6",
							Environments: "develop",
						},
					},
				},
				{
					Name: "7.6.1",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-elasticsearch",
							Tag:          "7.6.1",
							Build:        "5d1d0e6",
							Environments: "develop",
						},
					},
				},
			},
		},
		{
			Name:      "Couchbase",
			ShortName: "cbs",
			Templates: []Template{{
				Name:      "Small",
				SortOrder: 0,
				Resources: map[string]interface{}{
					"cpu":    "1",
					"memory": "4Gi",
					"disk":   "1Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "ftsRamsize",
						DefaultValue: "256",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "indexRamsize",
						DefaultValue: "256",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "ldapEnabled",
						DefaultValue: true,
						Type:         "boolean",
						Required:     true,
					},
					{
						Name:         "ramsize",
						DefaultValue: "512",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "restPassword",
						DefaultValue: "password",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "restUsername",
						DefaultValue: "username",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "services",
						DefaultValue: "data,index,query,fts",
						Type:         "string",
						Required:     true,
					},
				},
			}},
			Versions: []Version{
				{
					Name: "5.5.6",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-couchbase",
							Tag:          "5.5.6",
							Build:        "4a1ba8f",
							Environments: "develop",
						},
					},
				},
				{
					Name: "6.0.4",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-couchbase",
							Tag:          "6.0.4",
							Build:        "4a1ba8f",
							Environments: "develop",
						},
					},
				},
				{
					Name: "6.6.1",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-couchbase",
							Tag:          "6.6.1",
							Build:        "4a1ba8f",
							Environments: "develop",
						},
					},
				},
			},
		},
		{
			Name:      "Consul",
			ShortName: "con",
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
						Name:         "aclDefaultPolicy",
						DefaultValue: "deny",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "aclMasterToken",
						DefaultValue: "APPE0RDA2I8H",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "encryptionKey",
						DefaultValue: "pUqJrVyVRj5jsiYEkM/tFQYfWyJIv4s3XkvDwy7Cu5s=",
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
					Name: "1.1.0",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-consul",
							Tag:          "1.1.0",
							Build:        "24c6f02",
							Environments: "develop",
						},
					},
				},
				{
					Name: "1.4.0",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-consul",
							Tag:          "1.4.0",
							Build:        "24c6f02",
							Environments: "develop",
						},
					},
				},
				{
					Name: "1.7.2",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-consul",
							Tag:          "1.7.2",
							Build:        "24c6f02",
							Environments: "develop",
						},
					},
				},
				{
					Name: "1.8.0",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-consul",
							Tag:          "1.8.0",
							Build:        "24c6f02",
							Environments: "develop",
						},
					},
				},
			},
		},
		{
			Name:      "DatastaxEnterprise",
			ShortName: "dse",
			Templates: []Template{{
				Name:      "Small",
				SortOrder: 0,
				Resources: map[string]interface{}{
					"cpu":    "1",
					"memory": "1Gi",
					"disk":   "1Gi",
				},
				Configurations: []Configuration{
					{
						Name:         "CASSANDRA_AUTHENTICATOR",
						DefaultValue: "AllowAllAuthenticator",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "CASSANDRA_AUTHORIZER",
						DefaultValue: "AllowAllAuthorizer",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "graphEnabled",
						DefaultValue: true,
						Type:         "boolean",
						Required:     true,
					},
					{
						Name:         "heapNewSize",
						DefaultValue: "200M",
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
						Name:         "solrEnabled",
						DefaultValue: true,
						Type:         "boolean",
						Required:     true,
					},
					{
						Name:         "sparkEnabled",
						DefaultValue: true,
						Type:         "boolean",
						Required:     true,
					},
				},
			}},
			Versions: []Version{
				{
					Name: "6.0.14",
				},
				{
					Name: "6.8.11",
				},
			},
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
			}},
			Versions: []Version{
				{
					Name: "2.1.1",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-kafka",
							Tag:          "2.1.1",
							Build:        "d4a60bc",
							Environments: "develop",
						},
					},
				},
				{
					Name: "2.2.0",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-kafka",
							Tag:          "2.2.0",
							Build:        "d4a60bc",
							Environments: "develop",
						},
					},
				},
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
		},
		{
			Name:      "Mongodb",
			ShortName: "mdb",
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
						Name:         "CONTAINER_PORT",
						DefaultValue: "27017",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "MONGODB_ADMIN_PASSWORD",
						DefaultValue: "admin",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "MONGODB_CLUSTER_ROLE",
						DefaultValue: "shardsvr",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "MONGODB_DATABASE",
						DefaultValue: "mongodb",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "MONGODB_PASSWORD",
						DefaultValue: "password",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "MONGODB_REPLICA_NAME",
						DefaultValue: "rs1",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "MONGODB_USER",
						DefaultValue: "user",
						Type:         "string",
						Required:     true,
					},
					{
						Name:         "MONGODBDATA_DIR",
						DefaultValue: "/data",
						Type:         "string",
						Required:     true,
					},
				},
			}},
			Versions: []Version{
				{
					Name: "3.6.17",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-mongodb",
							Tag:          "3.6.17",
							Build:        "ba0773c",
							Environments: "develop",
						},
					},
				},
				{
					Name: "4.4.7",
					Images: []Image{
						{
							Registry:     "docker.io",
							Namespace:    "portworx",
							Name:         "pds-mongodb",
							Tag:          "4.4.7",
							Build:        "ba0773c",
							Environments: "develop",
						},
					},
				},
			},
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
		},
		{
			Name:      "Redis",
			ShortName: "red",
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
