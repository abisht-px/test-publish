package agent_installer

import (
	"fmt"

	"github.com/Masterminds/semver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func selectHelmChartVersions(contraints string, versions []string) ([]string, error) {
	c, err := semver.NewConstraint(contraints)
	if err != nil {
		return nil, err
	}

	var selectedVersions []string
	for _, v := range versions {
		if ver, err := semver.NewVersion(v); err == nil && c.Check(ver) {
			selectedVersions = append(selectedVersions, v)
		}
	}
	if len(selectedVersions) == 0 {
		return nil, fmt.Errorf("no version found with constraint %s", contraints)
	}
	return selectedVersions, nil
}

func restGetterFromKubeConfig(kubeconfig string) genericclioptions.RESTClientGetter {
	cf := genericclioptions.NewConfigFlags(true)
	cf.KubeConfig = &kubeconfig
	return cf
}
