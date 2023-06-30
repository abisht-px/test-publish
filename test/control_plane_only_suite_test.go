package test

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/controlplane"
)

type ControlPlaneTestSuite struct {
	suite.Suite
	ctx          context.Context
	ControlPlane *controlplane.ControlPlane
	config       controlPlaneEnvironment
}

func TestControlPlaneTestSuite(t *testing.T) {
	suite.Run(t, new(ControlPlaneTestSuite))
}

func (s *ControlPlaneTestSuite) SetupSuite() {
	// Try to load .env file from the root of the project.
	err := godotenv.Load("../.env")
	if err == nil {
		s.T().Log("successfully loaded .env file")
	}

	s.ctx = context.Background()

	config := mustHaveControlPlaneEnvVariables(s.T())
	s.config = config
	apiClient, err := api.NewPDSClient(context.Background(), config.ControlPlaneAPI, config.LoginCredentials)
	s.Require().NoError(err, "could not create Control Plane API client")

	s.ControlPlane = controlplane.New(apiClient)
	s.ControlPlane.MustInitializeTestData(s.ctx, s.T(),
		config.AccountName,
		config.TenantName,
		config.ProjectName,
		namePrefix)
}

func (s *ControlPlaneTestSuite) TearDownSuite() {
	s.ControlPlane.DeleteTestApplicationTemplates(s.ctx, s.T())
	s.ControlPlane.DeleteTestStorageOptions(s.ctx, s.T())
}
