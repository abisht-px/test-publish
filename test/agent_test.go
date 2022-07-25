package test

import (
	_ "github.com/stretchr/testify/suite" // Pin import to recognize suite tests: https://github.com/golang/vscode-go/issues/899

	agent_installer "github.com/portworx/pds-integration-test/internal/agent-installer"
)

// mustInstallAgent installs latest agent with chart values
func (s *PDSTestSuite) mustInstallAgent(env environment) {
	provider, err := agent_installer.NewHelmProvider()
	s.Require().NoError(err, "Cannot create agent installer provider.")

	versions, err := provider.Versions()
	s.Require().NoError(err, "Cannot get agent installer versions.")
	s.Require().NotEmpty(versions, "No agent installer versions found.")

	pdsAgentVersion := versions[0]
	installer, err := provider.Installer(pdsAgentVersion)
	s.Require().NoError(err, "Cannot get agent installer for version %s.", pdsAgentVersion)

	installArgs := map[string]interface{}{
		"tenantId":    s.testPDSTenantID,
		"bearerToken": s.testPDSAgentToken,
		// TODO: Get controlPlaneURL from apiClient.
		"apiEndpoint": env.controlPlaneAPI,
	}
	err = installer.Install(s.ctx, env.targetKubeconfig, installArgs)
	s.Require().NoError(err, "Cannot install agent for version %s.", pdsAgentVersion)
}
