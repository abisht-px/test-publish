package targetcluster

import (
	"context"

	"github.com/portworx/pds-integration-test/internal/helminstaller"
)

const (
	PDSChartReleaseName = "pds"
	PDSChartNamespace   = "pds-system"
)

var PDSOperators = map[string]struct {
	Name       string
	Deployment string
}{
	"backup":     {Name: "backup", Deployment: "pds-backup-controller-manager"},
	"deployment": {Name: "deployment", Deployment: "pds-deployment-controller-manager"},
	"target":     {Name: "target", Deployment: "pds-operator-target-controller-manager"},
}

type PDSChartConfig struct {
	Version               string
	TenantID              string
	Token                 string
	ControlPlaneAPI       string
	DeploymentTargetName  string
	DataServiceTLSEnabled bool
}

func (c PDSChartConfig) ToChartConfig() helminstaller.ChartConfig {
	chartConfig := helminstaller.ChartConfig{
		ReleaseName:        PDSChartReleaseName,
		VersionConstraints: c.Version,
		ChartValues: map[string]string{
			"tenantId":    c.TenantID,
			"bearerToken": c.Token,
			"apiEndpoint": c.ControlPlaneAPI,
			"clusterName": c.DeploymentTargetName,
		},
	}
	if c.DataServiceTLSEnabled {
		chartConfig.ChartValues["dataServiceTLSEnabled"] = "true"
	}
	return chartConfig
}

func (tc *TargetCluster) InstallPDSChart(ctx context.Context) error {
	installer, err := tc.PDSChartHelmProvider.Installer(tc.Kubeconfig, tc.PDSChartConfig.ToChartConfig())
	if err != nil {
		return err
	}
	return installer.Install(ctx)
}

// UpgradePDSChart runs helm upgrade with the configuration at tc.PDSChartConfig.
func (tc *TargetCluster) UpgradePDSChart(ctx context.Context) error {
	installer, err := tc.PDSChartHelmProvider.Installer(tc.Kubeconfig, tc.PDSChartConfig.ToChartConfig())
	if err != nil {
		return err
	}
	return installer.Upgrade(ctx)
}

func (tc *TargetCluster) UninstallPDSChart(ctx context.Context) error {
	installer, err := tc.PDSChartHelmProvider.Installer(tc.Kubeconfig, tc.PDSChartConfig.ToChartConfig())
	if err != nil {
		return err
	}
	return installer.Uninstall(ctx)
}
