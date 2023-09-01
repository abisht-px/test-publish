package dataservices_test

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/dataservices"
	"github.com/portworx/pds-integration-test/internal/kubernetes/psa"
)

const pdsSystemUsersCapabilityName = "pds_system_users"

var (
	latestCompatibleOnly = flag.Bool("latest-compatible-only", true, "Test only update to the latest compatible version.")
	skipBackups          = flag.Bool("skip-backups", false, "Skip tests related to backups.")
	skipBackupsMultinode = flag.Bool("skip-backups-multinode", true, "Skip tests related to backups which are run on multi-node data services.")

	// Common node counts per data service.
	// Tests can decide whether to use the low count, the high count, or both.
	// Assumption: counts are sorted in ascending order.
	commonNodeCounts = map[string][]int32{
		dataservices.Cassandra:     {1, 3},
		dataservices.Consul:        {1, 3},
		dataservices.Couchbase:     {1, 2},
		dataservices.ElasticSearch: {1, 3},
		dataservices.Kafka:         {1, 3},
		dataservices.MongoDB:       {1, 2},
		dataservices.MySQL:         {1, 2},
		dataservices.Postgres:      {1, 2},
		dataservices.RabbitMQ:      {1, 3},
		dataservices.Redis:         {1, 6},
		dataservices.SqlServer:     {1},
		dataservices.ZooKeeper:     {3},
	}
)

func (s *Dataservices) mustGetStatefulSetImageTag(dataServiceName string, imageVerstionString string) string {
	if dataServiceName != dataservices.SqlServer {
		return imageVerstionString
	}

	// Our SQLServer image version string is in the form of 2019-CU20-abc123,
	// but need to drop the last bit since the upstream MS image won't have this.
	substr := strings.Split(imageVerstionString, "-")
	s.Require().Equal(3, len(substr))
	return fmt.Sprintf("%s-%s", substr[0], substr[1])
}

func (s *Dataservices) updateTestImpl(ctx context.Context, t *testing.T, fromSpec, toSpec api.ShortDeploymentSpec) {
	deploymentID := s.controlPlane.MustDeployDeploymentSpec(ctx, t, &fromSpec)
	t.Cleanup(func() {
		s.controlPlane.MustRemoveDeployment(ctx, t, deploymentID)
		s.controlPlane.MustWaitForDeploymentRemoved(ctx, t, deploymentID)
		s.crossCluster.MustDeleteDeploymentVolumes(ctx, t, deploymentID)
	})

	// Create.
	s.controlPlane.MustWaitForDeploymentHealthy(ctx, t, deploymentID)
	s.crossCluster.MustWaitForDeploymentInitialized(ctx, t, deploymentID)
	s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
	s.crossCluster.MustWaitForLoadBalancerServicesReady(ctx, t, deploymentID)
	s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx, t, deploymentID)
	loadTestUser := s.crossCluster.MustGetLoadTestUser(ctx, t, deploymentID)
	s.crossCluster.MustRunLoadTestJobWithUser(ctx, t, deploymentID, loadTestUser)

	// Update.
	oldUpdateRevision := s.crossCluster.MustGetStatefulSetUpdateRevision(ctx, t, deploymentID)
	s.controlPlane.MustUpdateDeployment(ctx, t, deploymentID, &toSpec)
	s.crossCluster.MustWaitForStatefulSetChanged(ctx, t, deploymentID, oldUpdateRevision)
	s.crossCluster.MustWaitForStatefulSetReady(ctx, t, deploymentID)
	targetTag := s.mustGetStatefulSetImageTag(toSpec.DataServiceName, toSpec.ImageVersionString())
	s.crossCluster.MustWaitForStatefulSetImage(ctx, t, deploymentID, targetTag)
	s.crossCluster.MustWaitForLoadBalancerServicesReady(ctx, t, deploymentID)
	s.crossCluster.MustWaitForLoadBalancerHostsAccessibleIfNeeded(ctx, t, deploymentID)

	s.crossCluster.MustRunLoadTestJobWithUser(ctx, t, deploymentID, loadTestUser)
}

func getSupportedPSAPolicy(dataServiceName string) string {
	// https://pds.docs.portworx.com/concepts/pod-security-admission/#supported-security-levels-for-pds-resources
	switch dataServiceName {
	case dataservices.Cassandra:
		return psa.PSAPolicyPrivileged
	default:
		return psa.PSAPolicyRestricted
	}
}

// getPatchVersionNamePrefix returns the prefix of the version name that does not change on patch update.
func getPatchVersionNamePrefix(dataServiceName, versionName string) string {
	switch dataServiceName {
	case dataservices.SqlServer:
		return "2019-CU"
	case dataservices.ElasticSearch:
		// Elasticsearch's version lines are like 8.x.y
		firstDot := strings.Index(versionName, ".")
		return versionName[0 : firstDot+1]
	default:
		// For other data services the last number is used for patch updates.
		lastDot := strings.LastIndex(versionName, ".")
		return versionName[0 : lastDot+1]
	}
}

func filterImagesByVersionNamePrefix(images []pds.ModelsImage, versionNamePrefix string) []pds.ModelsImage {
	var filteredImages []pds.ModelsImage
	for _, image := range images {
		if strings.HasPrefix(*image.Tag, versionNamePrefix) {
			filteredImages = append(filteredImages, image)
		}
	}
	return filteredImages
}

func hasPDSSystemUsersCapability(image pds.ModelsImage) bool {
	v := image.Capabilities[pdsSystemUsersCapabilityName]
	return v != nil && v != ""
}
