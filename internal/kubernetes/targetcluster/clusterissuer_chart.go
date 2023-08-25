package targetcluster

import (
	"context"

	"github.com/portworx/pds-integration-test/internal/helminstaller"
)

const (
	CertManagerReleaseName = "cert-manager"
	CertManagerNamespace   = "cert-manager"
)

type CertManagerChartConfig struct {
	Version string
}

func (c CertManagerChartConfig) ToChartConfig() helminstaller.ChartConfig {
	return helminstaller.ChartConfig{
		ReleaseName:        CertManagerReleaseName,
		VersionConstraints: c.Version,
		ChartValues: map[string]string{
			"installCRDs": "true",
		},
	}
}

func (tc *TargetCluster) InstallCertManagerChart(ctx context.Context) error {
	installer, err := tc.CertManagerChartInstaller()
	if err != nil {
		return err
	}
	return installer.Install(ctx)
}

// UpgradeCertManagerChart runs helm upgrade with the configuration at tc.CertManagerChartConfig.
func (tc *TargetCluster) UpgradeCertManagerChart(ctx context.Context) error {
	installer, err := tc.CertManagerChartInstaller()
	if err != nil {
		return err
	}
	return installer.Upgrade(ctx)
}

func (tc *TargetCluster) UninstallCertManagerChart(ctx context.Context) error {
	installer, err := tc.CertManagerChartInstaller()
	if err != nil {
		return err
	}
	return installer.Uninstall(ctx)
}

func (tc *TargetCluster) CertManagerChartInstaller() (*helminstaller.InstallableHelm, error) {
	if tc.Kubeconfig != "" {
		return tc.CertManagerHelmProvider.Installer(tc.Kubeconfig, tc.CertManagerChartConfig.ToChartConfig())
	}

	return tc.CertManagerHelmProvider.InstallerFromRestCfg(
		tc.RestConfig,
		tc.CertManagerChartConfig.ToChartConfig(),
		CertManagerNamespace,
	)
}
