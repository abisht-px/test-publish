package dataservices

import (
	"time"

	"github.com/portworx/pds-integration-test/internal/wait"
)

// GetLongTimeoutFor returns a longer or shorter long timeout based on the deployment's node count.
func GetLongTimeoutFor(nodeCount int32) time.Duration {
	if nodeCount > 1 {
		return wait.VeryLongTimeout
	}
	return wait.LongTimeout
}
