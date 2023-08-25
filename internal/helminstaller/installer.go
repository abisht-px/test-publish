package helminstaller

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/portworx/pds-integration-test/internal/minihelm"
)

const (
	pdsRepoName  = "pds"
	pdsChartName = "pds-target"
	pdsRepoURL   = "https://portworx.github.io/pds-charts"

	certManagerRepoName  = "jetstack"
	certManagerChartName = "cert-manager"
	certManagerRepoURL   = "https://charts.jetstack.io"
)

type HelmArtifactProvider struct {
	client    minihelm.Client
	versions  []string
	chartName string
	repoName  string
}

type InstallableHelm struct {
	HelmArtifactProvider
	restGetter   genericclioptions.RESTClientGetter
	chartValues  string
	chartVersion string
	releaseName  string
}

func nullWriter(format string, v ...interface{}) {}

func NewHelmProviderPDS(namespace string) (*HelmArtifactProvider, error) {
	return newHelmProvider(pdsChartName, pdsRepoName, pdsRepoURL, namespace)
}

func NewHelmProviderCertManager(namespace string) (*HelmArtifactProvider, error) {
	return newHelmProvider(certManagerChartName, certManagerRepoName, certManagerRepoURL, namespace)
}

func newHelmProvider(chartName, repoName, repoURL, namespace string) (*HelmArtifactProvider, error) {
	var err error
	var client minihelm.Client
	if client, err = minihelm.New(&minihelm.ClientOptions{Namespace: namespace}); err != nil {
		return nil, fmt.Errorf("creating minihelm client: %w", err)
	}

	if !client.HasRepoWithNameAndURL(repoName, repoURL) {
		return nil, fmt.Errorf("repo %s not found", repoName)
	}

	err = client.UpdateRepo(repoName)
	if err != nil {
		return nil, fmt.Errorf("updating %s repo: %w", repoName, err)
	}

	versions, err := client.GetChartVersions(repoName, chartName)
	if err != nil {
		return nil, fmt.Errorf("getting versions of chart %s from repo %s: %w", pdsChartName, repoName, err)
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("repository %s does not have chart %s", repoName, pdsChartName)
	}

	return &HelmArtifactProvider{
		client:    client,
		versions:  versions,
		repoName:  repoName,
		chartName: chartName,
	}, nil
}

func (p *HelmArtifactProvider) InstallerFromRestCfg(cfg *rest.Config, chartConfig ChartConfig, namespace string) (*InstallableHelm, error) {
	matchingVersions, err := filterMatchingVersions(chartConfig.VersionConstraints, p.versions)
	if err != nil {
		return nil, err
	}

	return &InstallableHelm{
		HelmArtifactProvider: *p,
		chartVersion:         matchingVersions[0],
		restGetter: NewMemoryRESTClientGetter(
			cfg,
			WithPersistent(true),
			WithNamespace(namespace),
		),
		chartValues: chartConfig.CommaSeparatedChartVals(),
		releaseName: chartConfig.ReleaseName,
	}, nil
}

func (p *HelmArtifactProvider) Installer(kubeconfig string, chartConfig ChartConfig) (*InstallableHelm, error) {
	restClientGetter := genericclioptions.NewConfigFlags(true)
	restClientGetter.KubeConfig = &kubeconfig

	matchingVersions, err := filterMatchingVersions(chartConfig.VersionConstraints, p.versions)
	if err != nil {
		return nil, err
	}

	return &InstallableHelm{
		HelmArtifactProvider: *p,
		chartVersion:         matchingVersions[0],
		restGetter:           restClientGetter,
		chartValues:          chartConfig.CommaSeparatedChartVals(),
		releaseName:          chartConfig.ReleaseName,
	}, nil
}

func (i *InstallableHelm) Install(ctx context.Context) error {
	return i.client.InstallChart(ctx, i.restGetter, i.repoName, i.releaseName, i.chartName, i.chartVersion, i.chartValues, nullWriter)
}

func (i *InstallableHelm) Upgrade(ctx context.Context) error {
	return i.client.UpgradeChart(ctx, i.restGetter, i.repoName, i.releaseName, i.chartName, i.chartVersion, i.chartValues, nullWriter)
}

func (i *InstallableHelm) Version() string {
	return i.chartVersion
}

func (i *InstallableHelm) Uninstall(ctx context.Context) error {
	return i.client.UninstallChartVersion(ctx, i.restGetter, i.releaseName, nullWriter)
}
