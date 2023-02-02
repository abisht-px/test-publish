package helminstaller

import (
	"context"
	"fmt"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/portworx/pds-integration-test/internal/minihelm"
)

const (
	pdsRepoName    = "pds"
	pdsReleaseName = "pds"
	pdsChartName   = "pds-target"
	pdsRepoURL     = "https://portworx.github.io/pds-charts"
)

type ArtifactProviderHelmPDS struct {
	client   minihelm.Client
	versions []string
}

type InstallableHelmPDS struct {
	helmClient          minihelm.Client
	restGetter          genericclioptions.RESTClientGetter
	helmChartValsString string
	pdsRepoName         string
	pdsReleaseName      string
	pdsChartName        string
	pdsChartVersion     string
}

func nullWriter(format string, v ...interface{}) {}

func NewHelmProvider() (*ArtifactProviderHelmPDS, error) {
	var err error
	var client minihelm.Client
	if client, err = minihelm.New(); err != nil {
		return nil, err
	}

	if !client.HasRepoWithNameAndURL(pdsRepoName, pdsRepoURL) {
		return nil, fmt.Errorf("repo %s not found", pdsRepoName)
	}

	err = client.UpdateRepo(pdsRepoName)
	if err != nil {
		return nil, err
	}

	versions, err := client.GetChartVersions(pdsRepoName, pdsChartName)
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("repository %s does not have chart %s", pdsRepoName, pdsChartName)
	}

	return &ArtifactProviderHelmPDS{
		client:   client,
		versions: versions,
	}, nil
}

func (p *ArtifactProviderHelmPDS) Installer(kubeconfig string, pdsChartConfig *pdsChartConfig) (*InstallableHelmPDS, error) {
	restClientGetter := genericclioptions.NewConfigFlags(true)
	restClientGetter.KubeConfig = &kubeconfig

	matchingVersions, err := filterMatchingVersions(pdsChartConfig.VersionConstraints, p.versions)
	if err != nil {
		return nil, err
	}

	return &InstallableHelmPDS{
		pdsRepoName:         pdsRepoName,
		pdsReleaseName:      pdsReleaseName,
		pdsChartName:        pdsChartName,
		pdsChartVersion:     matchingVersions[0],
		helmClient:          p.client,
		restGetter:          restClientGetter,
		helmChartValsString: pdsChartConfig.CommaSeparatedChartVals(),
	}, nil
}

func (i *InstallableHelmPDS) Install(ctx context.Context) error {
	return i.helmClient.InstallChartVersion(ctx, i.restGetter, i.pdsRepoName, i.pdsChartName, i.pdsChartVersion, i.helmChartValsString, nullWriter)
}

func (i *InstallableHelmPDS) Version() string {
	return i.pdsChartVersion
}

func (i *InstallableHelmPDS) Uninstall(ctx context.Context) error {
	return i.helmClient.UninstallChartVersion(ctx, i.restGetter, i.pdsChartName, nullWriter)
}
