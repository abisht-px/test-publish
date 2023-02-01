package agent_installer

import (
	"context"
	"fmt"
	"strings"

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

type selectorHelmPDS struct {
	restGetter         genericclioptions.RESTClientGetter
	versionConstraints string
	helmChartVals      map[string]string
}

func nullWriter(format string, v ...interface{}) {}

func NewHelmProvider() (*ArtifactProviderHelmPDS, error) {
	return newArtifactProviderHelmPDS(pdsRepoName, pdsChartName)
}

func newArtifactProviderHelmPDS(pdsRepoName, pdsChartName string) (*ArtifactProviderHelmPDS, error) {
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

func (p *ArtifactProviderHelmPDS) Versions() ([]string, error) {
	return p.versions, nil
}

func (p *ArtifactProviderHelmPDS) Installer(selector Selector) (Installable, error) {
	helmSelector, ok := selector.(*selectorHelmPDS)
	if !ok {
		return nil, fmt.Errorf("PDS Helm provider can only receive Helm installable artefact, received not supported type %T", selector)
	}

	versions, err := p.Versions()
	if err != nil {
		return nil, err
	}

	selectedVersions, err := selectHelmChartVersions(selector.ConstraintsString(), versions)
	if err != nil {
		return nil, err
	}

	return &InstallableHelmPDS{
		pdsRepoName:         pdsRepoName,
		pdsReleaseName:      pdsReleaseName,
		pdsChartName:        pdsChartName,
		pdsChartVersion:     selectedVersions[0],
		helmClient:          p.client,
		restGetter:          helmSelector.restGetter,
		helmChartValsString: helmSelector.ChartVals(),
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

func (s *selectorHelmPDS) ChartVals() string {
	var b strings.Builder

	for keyString, valueString := range s.helmChartVals {
		fmt.Fprintf(&b, "%s=%s,", keyString, valueString)
	}
	return strings.TrimSuffix(b.String(), ",")
}

func (s *selectorHelmPDS) ConstraintsString() string {
	return s.versionConstraints
}
