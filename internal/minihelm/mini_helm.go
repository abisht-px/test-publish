package minihelm

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/strvals"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type DebugLog func(format string, v ...interface{})

type Client interface {
	HasRepoWithName(repoName string) bool
	HasRepoWithNameAndURL(repoName, url string) bool
	GetChartVersions(repoName, chartName string) ([]string, error)
	InstallChartVersion(ctx context.Context, restGetter genericclioptions.RESTClientGetter, repoName, chartName, ChartVersion, chartVals string, logger DebugLog) error
}

// miniHelm is a partial implementation of HelmCmd, w/o Helm storage mutating features (Add Chart, Add Repo etc.).
type miniHelm struct {
	settings  *cli.EnvSettings
	providers getter.Providers
	storage   *repo.File
}

func New() (Client, error) {
	client := miniHelm{}
	client.settings = cli.New()
	client.providers = getter.All(client.settings)

	repoFile := client.settings.RepositoryConfig

	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	bs, err := ioutil.ReadFile(repoFile)
	if err != nil {
		return nil, err
	}

	client.storage = &repo.File{}
	if err := yaml.Unmarshal(bs, client.storage); err != nil {
		return nil, err
	}

	return &client, nil
}

func (m *miniHelm) HasRepoWithName(repoName string) bool {
	return m.storage.Has(repoName)
}

func (m *miniHelm) HasRepoWithNameAndURL(repoName, url string) bool {
	entry := m.storage.Get(repoName)
	return entry != nil && entry.URL == url
}

func (m *miniHelm) GetChartVersions(repoName, chartName string) ([]string, error) {
	f := path.Join(m.settings.RepositoryCache, helmpath.CacheIndexFile(repoName))
	ind, err := repo.LoadIndexFile(f)
	if err != nil {
		return nil, err
	}

	ind.SortEntries()
	for name, chartVersions := range ind.Entries {
		if name == chartName {
			versions := make([]string, 0)
			for _, version := range chartVersions {
				versions = append(versions, version.Metadata.Version)
			}
			return versions, nil
		}
	}

	return nil, fmt.Errorf("chart %s not found", chartName)
}

func (m *miniHelm) InstallChartVersion(ctx context.Context, restGetter genericclioptions.RESTClientGetter, repoName, chartName, chartVersion, chartVals string, logger DebugLog) error {
	_, err := installPDSChartVersionWithContext(ctx, m.settings, restGetter, repoName, chartName, chartVersion, chartVals, action.DebugLog(logger))
	return err
}

func installPDSChartVersionWithContext(ctx context.Context, settings *cli.EnvSettings, restGetter genericclioptions.RESTClientGetter, repoName, chartName, chartVer, chartVals string, logger action.DebugLog) (*release.Release, error) {
	// When KUBECONFIG fails fallback to HELM namespace.
	namespace := settings.Namespace()
	if config, err := restGetter.ToRawKubeConfigLoader().RawConfig(); err == nil {
		namespace = config.Contexts[config.CurrentContext].Namespace
	}

	var actionConfig action.Configuration
	if err := actionConfig.Init(restGetter, namespace, os.Getenv("HELM_DRIVER"), logger); err != nil {
		return nil, err
	}
	client := action.NewInstall(&actionConfig)

	// Not using variable release name in the context of the usecase of the package.
	client.ReleaseName = chartName
	client.Version = chartVer
	if client.Version == "" {
		client.Version = ">0.0.0-0"
	}

	cp, err := client.ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", repoName, chartName), settings)
	if err != nil {
		return nil, err
	}

	providerGetters := getter.All(settings)
	valueOpts := &values.Options{}
	vals, err := valueOpts.MergeValues(providerGetters)
	if err != nil {
		return nil, err
	}

	if err := strvals.ParseInto(chartVals, vals); err != nil {
		return nil, err
	}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if !client.DependencyUpdate {
				return nil, err
			}
			man := &downloader.Manager{
				Out:              os.Stdout,
				ChartPath:        cp,
				Keyring:          client.ChartPathOptions.Keyring,
				SkipUpdate:       false,
				Getters:          providerGetters,
				RepositoryConfig: settings.RepositoryConfig,
				RepositoryCache:  settings.RepositoryCache,
			}
			if err := man.Update(); err != nil {
				return nil, err
			}
		}
	}

	client.Namespace = settings.Namespace()
	return client.RunWithContext(ctx, chartRequested, vals)
}
