package dataservices

import (
	"fmt"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"k8s.io/utils/pointer"
)

const (
	Cassandra     = "Cassandra"
	Couchbase     = "Couchbase"
	Kafka         = "Kafka"
	MongoDB       = "MongoDB Enterprise"
	MySQL         = "MySQL"
	Postgres      = "PostgreSQL"
	RabbitMQ      = "RabbitMQ"
	Redis         = "Redis"
	SqlServer     = "MS SQL Server"
	ZooKeeper     = "ZooKeeper"
	ElasticSearch = "Elasticsearch"
	Consul        = "Consul"
)

type TemplateSpec struct {
	ConfigurationTemplates map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest
	ResourceTemplates      map[string]pds.ControllersCreateResourceSettingsTemplateRequest
}

var (
	templateNameDefault = fmt.Sprintf("integration-test-default-%d", time.Now().Unix())
	templateNameSmall   = fmt.Sprintf("integration-test-small-%d", time.Now().Unix())
	templateNameMed     = fmt.Sprintf("integration-test-med-%d", time.Now().Unix())

	// When updating, please consider that the first element in each list is used by the setDeploymentDefaults function.
	TemplateSpecs = map[string]TemplateSpec{
		Cassandra: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
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
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.StringPtr(templateNameSmall),
					CpuRequest:     pointer.StringPtr("0.75"),
					CpuLimit:       pointer.StringPtr("1"),
					MemoryRequest:  pointer.StringPtr("1G"),
					MemoryLimit:    pointer.StringPtr("2G"),
					StorageRequest: pointer.StringPtr("5G"),
				},
				templateNameMed: {
					Name:           pointer.StringPtr(templateNameMed),
					CpuRequest:     pointer.StringPtr("1"),
					CpuLimit:       pointer.StringPtr("1.25"),
					MemoryRequest:  pointer.StringPtr("1500M"),
					MemoryLimit:    pointer.StringPtr("2500M"),
					StorageRequest: pointer.StringPtr("5G"),
				},
			},
		},
		Couchbase: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
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
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
		Consul: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
		Kafka: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
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
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
		MongoDB: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
		MySQL: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
		ElasticSearch: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
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
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
		Postgres: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
		RabbitMQ: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
		Redis: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
		SqlServer: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
		ZooKeeper: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.StringPtr(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
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
