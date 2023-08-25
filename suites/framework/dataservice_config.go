package framework

import (
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const activeVersionsYAML = `
dataservices:
  - name: Cassandra
    versions:
      - 4.1.2
      - 4.0.10
      - 3.11.15
      - 3.0.29
  - name: Consul
    versions:
      - 1.15.3
      - 1.14.7
  - name: Couchbase
    versions:
      - 7.1.1
      - 7.2.0
  - name: Elasticsearch
    versions:
      - 8.8.0
  - name: Kafka
    versions:
      - 3.4.1
      - 3.3.2
      - 3.2.3
      - 3.1.2
  - name: MongoDB Enterprise
    versions:
      - 6.0.6
  - name: MySQL
    versions:
      - 8.0.33
  - name: PostgreSQL
    versions:
      - "15.3"
      - "14.8"
      - "13.11"
      - "12.15"
      - "11.20"
  - name: RabbitMQ
    versions:
      - 3.11.16
      - 3.10.22
  - name: Redis
    versions:
      - 7.0.9
  - name: MS SQL Server
    versions:
      - 2019-CU20
  - name: ZooKeeper
    versions:
      - 3.8.1
      - 3.7.1
`

type Dataservice struct {
	Name     string   `yaml:"name"`
	Versions []string `yaml:"versions"`
}

type DSVersionMatrix struct {
	Dataservices []Dataservice `yaml:"dataservices"`
}

func (ds *DSVersionMatrix) GetVersions(dataservice string) []string {
	for _, each := range ds.Dataservices {
		if strings.EqualFold(each.Name, dataservice) {
			return each.Versions
		}
	}

	return nil
}

func (ds *DSVersionMatrix) HasDataservice(dataservice string) bool {
	for _, each := range ds.Dataservices {
		if strings.EqualFold(each.Name, dataservice) {
			return true
		}
	}

	return false
}

func NewDSVersionMatrixFromFlags() (DSVersionMatrix, error) {
	if DSVersionMatrixFile != "" {
		return NewDSVersionMatrixFromFile(DSVersionMatrixFile)
	}

	return DefaultDSVersionMatrix()
}

func NewDSVersionMatrixFromFile(paths ...string) (DSVersionMatrix, error) {
	filePath := path.Join(paths...)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return DSVersionMatrix{}, errors.Wrap(err, "read file")
	}

	return NewDSVersionMatrixFromBytes(data)
}

func NewDSVersionMatrixFromBytes(data []byte) (DSVersionMatrix, error) {
	obj := &DSVersionMatrix{}

	err := yaml.Unmarshal(data, obj)
	if err != nil {
		return DSVersionMatrix{}, errors.Wrap(err, "unmarshal data")
	}

	return *obj, nil
}

func DefaultDSVersionMatrix() (DSVersionMatrix, error) {
	return NewDSVersionMatrixFromBytes([]byte(activeVersionsYAML))
}

func NewDSVersionMatrixFromString(data string) (DSVersionMatrix, error) {
	dsList := []Dataservice{}

	for _, dsStr := range strings.Split(data, ";") {
		dsName, versionsStr, ok := strings.Cut(dsStr, "=")
		if !ok {
			return DSVersionMatrix{}, errors.New("invalid string value")
		}

		versions := strings.Split(versionsStr, ",")

		dsList = append(dsList, Dataservice{
			Name:     dsName,
			Versions: versions,
		})
	}

	return DSVersionMatrix{
		Dataservices: dsList,
	}, nil
}
