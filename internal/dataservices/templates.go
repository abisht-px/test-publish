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
					Name: pointer.String(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{
						{
							Key:   pointer.String("HEAP_NEWSIZE"),
							Value: pointer.String("400M"),
						},
						{
							Key:   pointer.String("MAX_HEAP_SIZE"),
							Value: pointer.String("1G"),
						},
						{
							Key:   pointer.String("PDS_VERBOSE_PROBE_CHECKS"),
							Value: pointer.String("1"),
						},
					},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("1"),
					CpuLimit:       pointer.String("1.25"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2000M"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("1.1"),
					CpuLimit:       pointer.String("1.35"),
					MemoryRequest:  pointer.String("1800M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
				},
			},
		},
		Couchbase: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name: pointer.String(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{
						{
							Key:   pointer.String("COUCHBASE_RAMSIZE"),
							Value: pointer.String("1024"),
						},
						{
							Key:   pointer.String("COUCHBASE_FTS_RAMSIZE"),
							Value: pointer.String("256"),
						},
						{
							Key:   pointer.String("COUCHBASE_INDEX_RAMSIZE"),
							Value: pointer.String("256"),
						},
					},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("1"),
					CpuLimit:       pointer.String("1.25"),
					MemoryRequest:  pointer.String("1750M"),
					MemoryLimit:    pointer.String("2750M"),
					StorageRequest: pointer.String("5G"),
				},
			},
		},
		Consul: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.String(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.5"),
					CpuLimit:       pointer.String("0.75"),
					MemoryRequest:  pointer.String("1G"),
					MemoryLimit:    pointer.String("2G"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
				},
			},
		},
		Kafka: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name: pointer.String(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{
						{
							Key:   pointer.String("heapSize"),
							Value: pointer.String("1500M"),
						},
					},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.5"),
					CpuLimit:       pointer.String("0.75"),
					MemoryRequest:  pointer.String("1G"),
					MemoryLimit:    pointer.String("2G"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
				},
			},
		},
		MongoDB: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.String(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.5"),
					CpuLimit:       pointer.String("0.75"),
					MemoryRequest:  pointer.String("1G"),
					MemoryLimit:    pointer.String("2G"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
				},
			},
		},
		MySQL: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.String(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.5"),
					CpuLimit:       pointer.String("0.75"),
					MemoryRequest:  pointer.String("1G"),
					MemoryLimit:    pointer.String("2G"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
				},
			},
		},
		ElasticSearch: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name: pointer.String(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{
						{
							Key:   pointer.String("HEAP_SIZE"),
							Value: pointer.String("1G"),
						},
					},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.5"),
					CpuLimit:       pointer.String("0.75"),
					MemoryRequest:  pointer.String("1G"),
					MemoryLimit:    pointer.String("2G"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
				},
			},
		},
		Postgres: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.String(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.5"),
					CpuLimit:       pointer.String("0.75"),
					MemoryRequest:  pointer.String("1G"),
					MemoryLimit:    pointer.String("2G"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
				},
			},
		},
		RabbitMQ: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.String(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.5"),
					CpuLimit:       pointer.String("0.75"),
					MemoryRequest:  pointer.String("1G"),
					MemoryLimit:    pointer.String("2G"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
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
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.5"),
					CpuLimit:       pointer.String("0.75"),
					MemoryRequest:  pointer.String("1G"),
					MemoryLimit:    pointer.String("2G"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
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
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.5"),
					CpuLimit:       pointer.String("0.75"),
					MemoryRequest:  pointer.String("1G"),
					MemoryLimit:    pointer.String("2G"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
				},
			},
		},
		ZooKeeper: {
			ConfigurationTemplates: map[string]pds.ControllersCreateApplicationConfigurationTemplateRequest{
				templateNameDefault: {
					Name:        pointer.String(templateNameDefault),
					ConfigItems: []pds.ModelsConfigItem{},
				},
			},
			ResourceTemplates: map[string]pds.ControllersCreateResourceSettingsTemplateRequest{
				templateNameSmall: {
					Name:           pointer.String(templateNameSmall),
					CpuRequest:     pointer.String("0.5"),
					CpuLimit:       pointer.String("0.75"),
					MemoryRequest:  pointer.String("1G"),
					MemoryLimit:    pointer.String("2G"),
					StorageRequest: pointer.String("5G"),
				},
				templateNameMed: {
					Name:           pointer.String(templateNameMed),
					CpuRequest:     pointer.String("0.75"),
					CpuLimit:       pointer.String("1"),
					MemoryRequest:  pointer.String("1500M"),
					MemoryLimit:    pointer.String("2500M"),
					StorageRequest: pointer.String("5G"),
				},
			},
		},
	}
)

func ToPluralName(dataServiceName string) string {
	switch dataServiceName {
	case Cassandra:
		return "cassandras"
	case Consul:
		return "consuls"
	case Couchbase:
		return "couchbases"
	case ElasticSearch:
		return "elasticsearches"
	case Kafka:
		return "kafkas"
	case MongoDB:
		return "mongodbs"
	case MySQL:
		return "mysqls"
	case Postgres:
		return "postgresqls"
	case RabbitMQ:
		return "rabbitmqs"
	case Redis:
		return "redis"
	case SqlServer:
		return "sqlservers"
	case ZooKeeper:
		return "zookeepers"
	}
	return ""
}
