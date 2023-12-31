package minihelm

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/pflag"
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

type Client interface {
	HasRepoWithNameAndURL(repoName, url string) bool
	UpdateRepo(repoName string) error
	GetChartVersions(repoName, chartName string) ([]string, error)
	InstallChart(ctx context.Context, restGetter genericclioptions.RESTClientGetter, repoName, releaseName, chartName, chartVersion, chartVals string, logger action.DebugLog) error
	UpgradeChart(ctx context.Context, restGetter genericclioptions.RESTClientGetter, repoName, releaseName, chartName, chartVersion, chartVals string, logger action.DebugLog) error
	UninstallChartVersion(ctx context.Context, restGetter genericclioptions.RESTClientGetter, releaseName string, logger action.DebugLog) error
}

// miniHelm is a partial implementation of HelmCmd, w/o Helm storage mutating features (Add Chart, Add Repo etc.).
type miniHelm struct {
	settings  *cli.EnvSettings
	providers getter.Providers
	storage   *repo.File
}

type ClientOptions struct {
	// The kubernetes namespace in which the client works.
	Namespace string
}

func New(options *ClientOptions) (Client, error) {
	client := miniHelm{}
	client.settings = cli.New()

	if options.Namespace != "" {
		pflags := pflag.NewFlagSet("", pflag.ContinueOnError)
		client.settings.AddFlags(pflags)
		err := pflags.Parse([]string{"-n", options.Namespace})
		if err != nil {
			return nil, err
		}
	}

	client.providers = getter.All(client.settings)

	repoFile := client.settings.RepositoryConfig

	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	bs, err := os.ReadFile(repoFile)
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

func (m *miniHelm) UpdateRepo(repoName string) error {
	entry := m.storage.Get(repoName)
	repo, err := repo.NewChartRepository(entry, getter.All(m.settings))
	if err != nil {
		return fmt.Errorf("creating repository for %s repo: %w", entry.Name, err)
	}
	_, err = repo.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("downloading index file for %s repo: %w", entry.Name, err)
	}
	return nil
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

func (m *miniHelm) InstallChart(ctx context.Context, restGetter genericclioptions.RESTClientGetter, repoName, releaseName, chartName, chartVersion, chartVals string, logger action.DebugLog) error {
	_, err := installChart(ctx, m.settings, restGetter, repoName, releaseName, chartName, chartVersion, chartVals, logger)
	return err
}

func (m *miniHelm) UpgradeChart(ctx context.Context, restGetter genericclioptions.RESTClientGetter, repoName, releaseName, chartName, chartVersion, chartVals string, logger action.DebugLog) error {
	_, err := upgradeChart(ctx, m.settings, restGetter, repoName, releaseName, chartName, chartVersion, chartVals, logger)
	return err
}

func (m *miniHelm) UninstallChartVersion(ctx context.Context, restGetter genericclioptions.RESTClientGetter, chartName string, logger action.DebugLog) error {
	// TODO (dbugrik): Add Context handling, Context is not naturally supported by helm v3 action yet
	_, err := uninstallPDSChartWithContext(m.settings, restGetter, chartName, logger)
	return err
}

func installChart(ctx context.Context, settings *cli.EnvSettings, restGetter genericclioptions.RESTClientGetter, repoName, releaseName, chartName, chartVer, chartVals string, logger action.DebugLog) (*release.Release, error) {
	namespace := setGetterNamespace(settings, restGetter)

	var actionConfig action.Configuration
	if err := actionConfig.Init(restGetter, namespace, os.Getenv("HELM_DRIVER"), logger); err != nil {
		return nil, err
	}
	client := action.NewInstall(&actionConfig)

	client.ReleaseName = releaseName
	client.Version = chartVer
	if client.Version == "" {
		client.Version = ">0.0.0-0"
	}
	client.CreateNamespace = true

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

func upgradeChart(ctx context.Context, settings *cli.EnvSettings, restGetter genericclioptions.RESTClientGetter, repoName, releaseName, chartName, chartVer, chartVals string, logger action.DebugLog) (*release.Release, error) {
	namespace := setGetterNamespace(settings, restGetter)

	var actionConfig action.Configuration
	if err := actionConfig.Init(restGetter, namespace, os.Getenv("HELM_DRIVER"), logger); err != nil {
		return nil, err
	}
	client := action.NewUpgrade(&actionConfig)
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
	return client.RunWithContext(ctx, releaseName, chartRequested, vals)
}

func uninstallPDSChartWithContext(settings *cli.EnvSettings, restGetter genericclioptions.RESTClientGetter, releaseName string, logger action.DebugLog) (*release.UninstallReleaseResponse, error) {
	namespace := setGetterNamespace(settings, restGetter)

	var actionConfig action.Configuration
	if err := actionConfig.Init(restGetter, namespace, os.Getenv("HELM_DRIVER"), logger); err != nil {
		return nil, err
	}

	client := action.NewUninstall(&actionConfig)
	return client.Run(releaseName)
}

func setGetterNamespace(settings *cli.EnvSettings, restGetter genericclioptions.RESTClientGetter) string {
	// HACK: Update getter's (a.k.a. client from KUBECONFIG) namespace.
	namespace := settings.Namespace()
	if config, err := restGetter.ToRawKubeConfigLoader().RawConfig(); err == nil {
		currCtxCfg := config.Contexts[config.CurrentContext]
		if currCtxCfg != nil {
			currCtxCfg.Namespace = namespace
		}
	}
	return namespace
}
