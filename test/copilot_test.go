package test

import (
	"net/http"

	"github.com/portworx/pds-integration-test/internal/api"
)

func (s *PDSTestSuite) TestCopilotSearch_SanityCheck() {

	// TODO: compare response from copilot, currently not possible because responses are nondeterministic
	testcases := []map[string]string{
		{
			"dataservice": "Cassandra",
			"query":       "list all employees whose age is greater than 25",
			"response":    "",
		},
		{
			"dataservice": "Couchbase",
			"query":       "list all employees whose age is greater than 25",
			"response":    "",
		},
		{
			"dataservice": "DataStaxEnterprise",
			"query":       "list all employees whose age is greater than 25",
			"response":    "",
		},
		{
			"dataservice": "Consul",
			"query":       "list all employees whose age is greater than 25",
			"response":    "",
		},
		{
			"dataservice": "MongoDB Enterprise",
			"query":       "list all employees whose age is greater than 25",
			"response":    "",
		},
		{
			"dataservice": "PostgreSQL",
			"query":       "list all employees whose age is greater than 25",
			"response":    "",
		},
		{
			"dataservice": "Kafka",
			"query":       "list all employees whose age is greater than 25",
			"response":    "",
		},
		{
			"dataservice": "Redis",
			"query":       "list all employees whose age is greater than 25",
			"response":    "",
		},
		{
			"dataservice": "ZooKeeper",
			"query":       "Check if the znode located at path '/my_node' exists",
			"response":    "",
		},
		{
			"dataservice": "MS SQL Server",
			"query":       "list all employees whose age is greater than 25",
			"response":    "",
		},
		{
			"dataservice": "RabbitMQ",
			"query":       "Create a queue named my_queue",
			"response":    "",
		},
		{
			"dataservice": "Elasticsearch",
			"query":       "Count the number of documents in the 'my_index' index",
			"response":    "",
		},
		{
			"dataservice": "MySQL",
			"query":       "list all employees whose age is greater than 25",
			"response":    "",
		},
	}

	dataservices := s.controlPlane.MustGetDataServicesByName(s.ctx, s.T())

	for _, tc := range testcases {
		copilotResp, resp, err := s.controlPlane.PerformCopilotQuery(
			s.ctx, s.T(), *dataservices[tc["dataservice"]].Id, tc["query"],
		)
		api.RequireNoError(s.T(), resp, err)
		s.Require().True(len(*copilotResp.Response) > 0)
	}

}

func (s *PDSTestSuite) TestCopilotSearch_InvalidQueries() {

	testcases := []map[string]string{
		{
			"dataservice": "Cassandra",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "Couchbase",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "DataStaxEnterprise",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "Consul",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "MongoDB Enterprise",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "PostgreSQL",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "Kafka",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "Redis",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "ZooKeeper",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "MS SQL Server",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "RabbitMQ",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "Elasticsearch",
			"query":       "ABCDEFGHJ",
		},
		{
			"dataservice": "MySQL",
			"query":       "ABCDEFGHJ",
		},
	}

	dataservices := s.controlPlane.MustGetDataServicesByName(s.ctx, s.T())

	for _, tc := range testcases {
		copilotResp, resp, err := s.controlPlane.PerformCopilotQuery(
			s.ctx, s.T(), *dataservices[tc["dataservice"]].Id, tc["query"],
		)
		s.Require().Error(err, "Provided query is not a valid dataservice query.")
		s.Require().Equal(http.StatusUnprocessableEntity, resp.StatusCode)
		s.Require().Nil(copilotResp)
	}

}
