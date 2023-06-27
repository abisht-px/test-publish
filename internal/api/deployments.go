package api

import (
	"fmt"
)

// PDSImageReferenceSpec is a summary specification of every image version with enough information to pass to a deployment call.
type PDSImageReferenceSpec struct {
	DataServiceName   string
	DataServiceID     string
	VersionID         string
	ImageVersionBuild string
	ImageVersionTag   string
	ImageID           string
}

// ShortDeploymentSpec is a shortened specification of Deployment.
// NOTE: Using only ImageVersionBuild should be sufficient but not 100% guaranteed uniqueness though.
type ShortDeploymentSpec struct {
	DataServiceName              string `yaml:"service_name"`
	ImageVersionTag              string `yaml:"image_version_tag"`
	ImageVersionBuild            string `yaml:"image_version_build"`
	AppConfigTemplateName        string `yaml:"app_config_template_name"`
	BackupPolicyname             string `yaml:"backup_policy_name"`
	StorageOptionName            string `yaml:"storage_option_name"`
	ResourceSettingsTemplateName string `yaml:"resource_settings_template_name"`
	ServiceType                  string `yaml:"service_type"`
	NamePrefix                   string `yaml:"name_prefix"`
	CRDNamePlural                string `yaml:"crd_name_plurals"`
	NodeCount                    int32  `yaml:"node_count"`
	BackupTargetName             string `yaml:"backup_target_name"`
}

func (d ShortDeploymentSpec) ImageVersionString() string {
	if d.ImageVersionTag != "" {
		if d.ImageVersionBuild != "" {
			return fmt.Sprintf("%s-%s", d.ImageVersionTag, d.ImageVersionBuild)
		} else {
			return d.ImageVersionTag
		}
	}
	return d.ImageVersionBuild
}
