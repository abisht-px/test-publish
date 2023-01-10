package test

import (
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"k8s.io/utils/pointer"
)

const (
	dbCassandra     = "Cassandra"
	dbCouchbase     = "Couchbase"
	dbKafka         = "Kafka"
	dbMongoDB       = "MongoDB Enterprise"
	dbMySQL         = "MySQL"
	dbPostgres      = "PostgreSQL"
	dbRabbitMQ      = "RabbitMQ"
	dbRedis         = "Redis"
	dbZooKeeper     = "ZooKeeper"
	dbElasticSearch = "Elasticsearch"
	dbConsul        = "Consul"
)

type dataServiceTemplateSpec struct {
	configurationTemplate pds.ControllersCreateApplicationConfigurationTemplatesRequest
	resourceTemplate      pds.ControllersCreateResourceSettingsTemplatesRequest
}

var (
	dataServiceTemplatesSpec = map[string]dataServiceTemplateSpec{
		dbCassandra: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{
					{
						Key:   pointer.StringPtr("CASSANDRA_AUTHORIZER"),
						Value: pointer.StringPtr("AllowAllAuthorizer"),
					},
					{
						Key:   pointer.StringPtr("CASSANDRA_AUTHENTICATOR"),
						Value: pointer.StringPtr("AllowAllAuthenticator"),
					},
					{
						Key:   pointer.StringPtr("HEAP_NEWSIZE"),
						Value: pointer.StringPtr("400M"),
					},
					{
						Key:   pointer.StringPtr("MAX_HEAP_SIZE"),
						Value: pointer.StringPtr("1G"),
					},
					{
						Key:   pointer.StringPtr("CASSANDRA_RACK"),
						Value: pointer.StringPtr("rack1"),
					},
					{
						Key:   pointer.StringPtr("CASSANDRA_DC"),
						Value: pointer.StringPtr("dc1"),
					},
				},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("1"),
				CpuLimit:       pointer.StringPtr("2"),
				MemoryRequest:  pointer.StringPtr("2G"),
				MemoryLimit:    pointer.StringPtr("4G"),
				StorageRequest: pointer.StringPtr("10G"),
			},
		},
		dbCouchbase: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{
					{
						Key:   pointer.StringPtr("COUCHBASE_RAMSIZE"),
						Value: pointer.StringPtr("2048"),
					},
					{
						Key:   pointer.StringPtr("COUCHBASE_FTS_RAMSIZE"),
						Value: pointer.StringPtr("256"),
					},
					{
						Key:   pointer.StringPtr("COUCHBASE_INDEX_RAMSIZE"),
						Value: pointer.StringPtr("256"),
					},
				},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("1"),
				CpuLimit:       pointer.StringPtr("2"),
				MemoryRequest:  pointer.StringPtr("2G"),
				MemoryLimit:    pointer.StringPtr("4G"),
				StorageRequest: pointer.StringPtr("10G"),
			},
		},
		dbConsul: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("0.5"),
				CpuLimit:       pointer.StringPtr("1"),
				MemoryRequest:  pointer.StringPtr("1G"),
				MemoryLimit:    pointer.StringPtr("2G"),
				StorageRequest: pointer.StringPtr("5G"),
			},
		},
		dbKafka: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{
					{
						Key:   pointer.StringPtr("heapSize"),
						Value: pointer.StringPtr("1500M"),
					},
				},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("0.5"),
				CpuLimit:       pointer.StringPtr("1"),
				MemoryRequest:  pointer.StringPtr("1G"),
				MemoryLimit:    pointer.StringPtr("2G"),
				StorageRequest: pointer.StringPtr("5G"),
			},
		},
		dbMongoDB: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("0.4"),
				CpuLimit:       pointer.StringPtr("0.5"),
				MemoryRequest:  pointer.StringPtr("800M"),
				MemoryLimit:    pointer.StringPtr("1G"),
				StorageRequest: pointer.StringPtr("5G"),
			},
		},
		dbMySQL: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("0.4"),
				CpuLimit:       pointer.StringPtr("0.5"),
				MemoryRequest:  pointer.StringPtr("800M"),
				MemoryLimit:    pointer.StringPtr("1G"),
				StorageRequest: pointer.StringPtr("5G"),
			},
		},
		dbElasticSearch: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{
					{
						Key:   pointer.StringPtr("HEAP_SIZE"),
						Value: pointer.StringPtr("2G"),
					},
				},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("1"),
				CpuLimit:       pointer.StringPtr("2"),
				MemoryRequest:  pointer.StringPtr("2G"),
				MemoryLimit:    pointer.StringPtr("4G"),
				StorageRequest: pointer.StringPtr("5G"),
			},
		},
		dbPostgres: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{
					{
						Key:   pointer.StringPtr("PG_DATABASE"),
						Value: pointer.StringPtr("pds"),
					},
				},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("0.5"),
				CpuLimit:       pointer.StringPtr("1"),
				MemoryRequest:  pointer.StringPtr("500M"),
				MemoryLimit:    pointer.StringPtr("1G"),
				StorageRequest: pointer.StringPtr("5G"),
			},
		},
		dbRabbitMQ: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{
					{
						Key:   pointer.StringPtr("RABBITMQ_DEFAULT_USER"),
						Value: pointer.StringPtr("pds"),
					},
					{
						Key:   pointer.StringPtr("DEFAULT_VHOST"),
						Value: pointer.StringPtr("/"),
					},
				},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("0.5"),
				CpuLimit:       pointer.StringPtr("1"),
				MemoryRequest:  pointer.StringPtr("300M"),
				MemoryLimit:    pointer.StringPtr("400M"),
				StorageRequest: pointer.StringPtr("5G"),
			},
		},
		dbRedis: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("0.5"),
				CpuLimit:       pointer.StringPtr("1"),
				MemoryRequest:  pointer.StringPtr("1G"),
				MemoryLimit:    pointer.StringPtr("2G"),
				StorageRequest: pointer.StringPtr("5G"),
			},
		},
		dbZooKeeper: {
			configurationTemplate: pds.ControllersCreateApplicationConfigurationTemplatesRequest{
				ConfigItems: []pds.ModelsConfigItem{
					{
						Key:   pointer.StringPtr("ZOO_4LW_COMMANDS_WHITELIST"),
						Value: pointer.StringPtr("*"),
					},
				},
			},
			resourceTemplate: pds.ControllersCreateResourceSettingsTemplatesRequest{
				CpuRequest:     pointer.StringPtr("0.5"),
				CpuLimit:       pointer.StringPtr("1"),
				MemoryRequest:  pointer.StringPtr("1G"),
				MemoryLimit:    pointer.StringPtr("2G"),
				StorageRequest: pointer.StringPtr("5G"),
			},
		},
	}
)
