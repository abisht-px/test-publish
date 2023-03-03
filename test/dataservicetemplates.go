package test

import (
	"fmt"
	"time"

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
	configurationTemplates map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest
	resourceTemplates      map[string]pds.ControllersCreateResourceSettingsTemplateRequest
}

var (
	templateNameDefault = fmt.Sprintf("integration-test-default-%d", time.Now().Unix())
	templateNameSmall   = fmt.Sprintf("integration-test-small-%d", time.Now().Unix())
	templateNameMed     = fmt.Sprintf("integration-test-med-%d", time.Now().Unix())

	// When updating, please consider that the first element in each list is used by the setDeploymentDefaults function.
	dataServiceTemplatesSpec = map[string]dataServiceTemplateSpec{
		dbCassandra: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name: pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{
						{
							Key:   pointer.StringPtr("HEAP_NEWSIZE"),
							Value: pointer.StringPtr("400M"),
						},
						{
							Key:   pointer.StringPtr("MAX_HEAP_SIZE"),
							Value: pointer.StringPtr("1G"),
						},
					},
				},
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.5"),
					CpuLimit:       pointer.StringPtr("0.75"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		dbCouchbase: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name: pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{
						{
							Key:   pointer.StringPtr("COUCHBASE_RAMSIZE"),
							Value: pointer.StringPtr("1024"),
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
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("1"),
					CpuLimit:       pointer.StringPtr("1.25"),
					MemoryRequest:  pointer.StringPtr("1750M"),
					MemoryLimit:    pointer.StringPtr("2750M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		dbConsul: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.5"),
					CpuLimit:       pointer.StringPtr("0.75"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		dbKafka: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name: pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{
						{
							Key:   pointer.StringPtr("heapSize"),
							Value: pointer.StringPtr("1500M"),
						},
					},
				},
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.5"),
					CpuLimit:       pointer.StringPtr("0.75"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		dbMongoDB: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.5"),
					CpuLimit:       pointer.StringPtr("0.75"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		dbMySQL: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.5"),
					CpuLimit:       pointer.StringPtr("0.75"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		dbElasticSearch: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name: pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{
						{
							Key:   pointer.StringPtr("HEAP_SIZE"),
							Value: pointer.StringPtr("1G"),
						},
					},
				},
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.5"),
					CpuLimit:       pointer.StringPtr("0.75"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		dbPostgres: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.5"),
					CpuLimit:       pointer.StringPtr("0.75"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		dbRabbitMQ: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.5"),
					CpuLimit:       pointer.StringPtr("0.75"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		dbRedis: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.5"),
					CpuLimit:       pointer.StringPtr("0.75"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		dbZooKeeper: {
			configurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			resourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.5"),
					CpuLimit:       pointer.StringPtr("0.75"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
	}
)
