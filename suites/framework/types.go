package framework

import "github.com/portworx/pds-integration-test/internal/controlplane"

type BackupTargetConfig struct {
	Bucket      string
	Region      string
	Credentials controlplane.BackupCredentials
}
