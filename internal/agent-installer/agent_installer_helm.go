package agent_installer

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/portworx/pds-integration-test/internal/minihelm"
)

const (
	pdsRepoName  = "pds"
	pdsChartName = "pds-target"
	pdsRepoURL   = "https://portworx.github.io/pds-charts"
)

type ArtifactProviderHelmPDS struct {
	client   minihelm.Client
	versions []string
}

type InstallableHelmPDS struct {
	client          minihelm.Client
	pdsRepoName     string
	pdsChartName    string
	pdsChartVersion string
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

	var versions []string
	versions, err = client.GetChartVersions(pdsRepoName, pdsChartName)
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

func (p *ArtifactProviderHelmPDS) Installer(version string) (Installable, error) {
	return &InstallableHelmPDS{
		pdsRepoName:     pdsRepoName,
		pdsChartName:    pdsChartName,
		pdsChartVersion: version,
		client:          p.client,
	}, nil
}

func (i *InstallableHelmPDS) Install(ctx context.Context, kubeconfig string, args map[string]interface{}) error {
	var kv string
	if args != nil {
		for k, v := range args {
			kv += fmt.Sprintf("%s=%v,", k, v)
		}
		kv = strings.TrimSuffix(kv, ",")
	}

	cf := genericclioptions.NewConfigFlags(true)
	cf.KubeConfig = &kubeconfig

	return i.client.InstallChartVersion(ctx, cf, i.pdsRepoName, i.pdsChartName, i.pdsChartVersion, kv, nullWriter)
}

func (i *InstallableHelmPDS) Version() string {
	return i.pdsChartVersion
}
