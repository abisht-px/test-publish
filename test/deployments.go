package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// PDSImageReferenceSpec is a summary specification of every image version with enough information to pass to a deployment call.
type PDSImageReferenceSpec struct {
	ServiceName       string
	DataServiceID     string
	VersionID         string
	ImageVersionBuild string
	ImageVersionTag   string
	ImageID           string
}

// ShortDeploymentSpec is a shortened specification of Deployment.
// NOTE: Using only ImageVersionBuild should be sufficient but not 100% guaranteed uniqueness though.
type ShortDeploymentSpec struct {
	ServiceName                  string `yaml:"service_name"`
	ImageVersionTag              string `yaml:"image_version_tag"`
	ImageVersionBuild            string `yaml:"image_version_build"`
	AppConfigTemplateName        string `yaml:"app_config_template_name"`
	BackupPolicyname             string `yaml:"backup_policy_name"`
	StorageOptionName            string `yaml:"storage_option_name"`
	ResourceSettingsTemplateName string `yaml:"resource_settings_template_name"`
	ServiceType                  string `yaml:"service_type"`
	NamePrefix                   string `yaml:"name_prefix"`
	NodeCount                    int    `yaml:"node_count"`
}

func (d ShortDeploymentSpec) getImageVersionString() string {
	if d.ImageVersionTag != "" {
		if d.ImageVersionBuild != "" {
			return fmt.Sprintf("%s-%s", d.ImageVersionTag, d.ImageVersionBuild)
		} else {
			return d.ImageVersionTag
		}
	}
	return d.ImageVersionBuild
}

type PDSDeploymentSpecID int

// mustLoadShortDeploymentSpecMap loads yaml flow encoded ShortDeploymentSpec from all environment variables with suffix envShortDeploymentSpecPrefix
// example of flow encoded single line: {key: value, key: val}
// each record may have additional attribute 'spec_id' which is used as SpecID key, otherwise env suffix is used
// NOTE: All environment variables must live together in a single environment block, which itself has a limit of 32767 characters.
func mustLoadShortDeploymentSpecMap(t *testing.T) map[PDSDeploymentSpecID]ShortDeploymentSpec {
	t.Helper()
	envVarList := mustGetEnvList(t, envShortDeploymentSpecPrefix)
	deployments := make(map[PDSDeploymentSpecID]ShortDeploymentSpec, len(envVarList))

	type ShortDeploymentSpecWithId struct {
		ShortDeploymentSpec `yaml:",inline"`
		SpecID              *PDSDeploymentSpecID `yaml:"spec_id"`
	}

	for idx, item := range envVarList {
		var d ShortDeploymentSpecWithId
		err := yaml.Unmarshal([]byte(item), &d)
		require.NoError(t, err, "failed to unmarshal deployment spec %d", idx)

		key := PDSDeploymentSpecID(idx)
		if d.SpecID != nil {
			key = *d.SpecID
		}
		require.NotContains(t, deployments, key, "duplicate deployment spec %d", key)

		deployments[key] = d.ShortDeploymentSpec
	}
	return deployments
}
