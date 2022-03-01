package test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.anim.dreamworks.com/DreamCloud/stella-api/api/models"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/suite"

	client "github.com/portworx/pds-integration-test/test/client"
	cluster "github.com/portworx/pds-integration-test/test/cluster"
	"github.com/portworx/pds-integration-test/test/color"
)

type PDSTestSuite struct {
	suite.Suite
	ControlPlane    *cluster.ControlPlane
	TargetCluster   *cluster.Target
	TestEnvironment *models.Environment
	ctx             context.Context
	apiClient       *pds.APIClient
}

func (s *PDSTestSuite) SetupSuite() {
	// Perform basic setup with sanity checks.
	env := mustHaveEnvVariables(s.T())
	s.mustHaveControlPlane(env)
	endpointUrl, err := url.Parse(env.controlPlaneAPI)
	if err != nil {
		s.T().Errorf("Unable to access the URL: %s", env.controlPlaneAPI)
	}
	apiConf := pds.NewConfiguration()
	apiConf.Host = endpointUrl.Host
	apiConf.Scheme = endpointUrl.Scheme
	s.ctx = context.WithValue(context.Background(), pds.ContextAPIKeys, map[string]pds.APIKey{"ApiKeyAuth": {Key: env.bearerToken, Prefix: "Bearer"}})
	s.apiClient = pds.NewAPIClient(apiConf)

}

func (s *PDSTestSuite) AfterTest(suiteName, testName string) {
	if s.T().Failed() {
		s.T().Log(color.Red(fmt.Sprintf("Failed test %s:", testName)))
		s.ControlPlane.LogStatus(s.T())
		s.TargetCluster.LogStatus(s.T(), s.TestEnvironment.Name)
	}
}

func (s *PDSTestSuite) mustHaveControlPlane(env environment) {
	apiClient := client.NewAPI(env.controlPlaneAPI)
	kubeContext := env.controlPlaneKubeconfig
	s.ControlPlane = cluster.NewControlPlane(apiClient, kubeContext)
}

func TestPDSSuite(t *testing.T) {
	suite.Run(t, new(PDSTestSuite))
}
