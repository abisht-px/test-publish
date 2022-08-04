package test

import (
	_ "github.com/stretchr/testify/suite" // Pin import to recognize suite tests: https://github.com/golang/vscode-go/issues/899

	agent_installer "github.com/portworx/pds-integration-test/internal/agent-installer"
)

func (s *PDSTestSuite) mustInstallAgent(env environment) {
	provider, err := agent_installer.NewHelmProvider()
	s.Require().NoError(err, "Cannot create agent installer provider.")

	helmSelectorAgent14, err := agent_installer.NewSelectorHelmPDS14(env.targetKubeconfig, s.testPDSTenantID, s.testPDSAgentToken, env.controlPlaneAPI)
	s.Require().NoError(err, "Cannot create agent installer selector.")

	installer, err := provider.Installer(helmSelectorAgent14)
	s.Require().NoError(err, "Cannot get agent installer for version selector %s.", helmSelectorAgent14.ConstraintsString())

	err = installer.Install(s.ctx)
	s.Require().NoError(err, "Cannot install agent for version %s selector.", helmSelectorAgent14.ConstraintsString())
	s.pdsAgentInstallable = installer
}

func (s *PDSTestSuite) mustUninstallAgent() {
	err := s.pdsAgentInstallable.Uninstall(s.ctx)
	s.Require().NoError(err)
}
