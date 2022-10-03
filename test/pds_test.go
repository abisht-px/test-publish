package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"

	agent_installer "github.com/portworx/pds-integration-test/internal/agent-installer"
	"github.com/portworx/pds-integration-test/internal/loadgen"
	"github.com/portworx/pds-integration-test/internal/loadgen/postgresql"
	"github.com/portworx/pds-integration-test/test/auth"
	"github.com/portworx/pds-integration-test/test/cluster"
)

const (
	waiterRetryInterval                          = time.Second * 10
	waiterDeploymentTargetNameExistsTimeout      = time.Second * 30
	waiterNamespaceExistsTimeout                 = time.Second * 30
	waiterDeploymentTargetStatusHealthyTimeout   = time.Second * 120
	waiterDeploymentTargetStatusUnhealthyTimeout = time.Second * 300
	waiterDeploymentStatusHealthyTimeout         = time.Second * 300
	waiterBackupStatusSucceededTimeout           = time.Second * 300
	waiterDeploymentStatusRemovedTimeout         = time.Second * 300
	waiterLoadTestJobFinishedTimeout             = time.Second * 300
)

var (
	namePrefix = fmt.Sprintf("integration-test-%d", time.Now().Unix())
)

type PDSTestSuite struct {
	suite.Suite
	ctx       context.Context
	startTime time.Time

	targetCluster              *cluster.TargetCluster
	targetClusterKubeconfig    string
	apiClient                  *pds.APIClient
	pdsAgentInstallable        agent_installer.Installable
	testPDSAccountID           string
	testPDSTenantID            string
	testPDSProjectID           string
	testPDSNamespaceID         string
	testPDSDeploymentTargetID  string
	testPDSServiceAccountID    string
	testPDSBackupCredentialsID string
	testPDSBackupTargetID      string
	testPDSAgentToken          string
	shortDeploymentSpecMap     map[PDSDeploymentSpecID]ShortDeploymentSpec
	imageVersionSpecList       []PDSImageReferenceSpec
}

func TestPDSSuite(t *testing.T) {
	suite.Run(t, new(PDSTestSuite))
}

func (s *PDSTestSuite) SetupSuite() {
	s.startTime = time.Now()
	s.ctx = context.Background()

	// Perform basic setup with sanity checks.
	env := mustHaveEnvVariables(s.T())
	s.targetClusterKubeconfig = env.targetKubeconfig
	s.shortDeploymentSpecMap = mustLoadShortDeploymentSpecMap(s.T())
	s.mustHaveTargetCluster(env)
	s.mustHaveTargetClusterNamespaces(env)
	s.mustHaveAPIClient(env)
	s.mustHavePDStestAccount(env)
	s.mustHavePDStestTenant(env)
	s.mustHavePDStestProject(env)
	s.mustHavePDStestServiceAccount(env)
	s.mustHavePDStestAgentToken(env)

	s.mustLoadImageVersions()
	s.mustInstallAgent(env)
	s.mustHavePDStestDeploymentTarget(env)
	s.mustHavePDStestNamespace(env)
	s.mustEnsureS3BackupCredentials(env)
	s.mustEnsureS3BackupTarget(env)
}

func (s *PDSTestSuite) TearDownSuite() {
	env := mustHaveEnvVariables(s.T())
	s.mustUninstallAgent(env)
	s.mustDeleteBackupTarget()
	s.mustDeleteBackupCredentials()
	s.mustDeletePDStestDeploymentTarget()
	if s.T().Failed() {
		s.targetCluster.LogComponents(s.T(), s.ctx, s.startTime)
	}
}

// mustHavePDStestAccount finds PDS account with name set in envrionment and stores its ID as "Test PDS Account".
func (s *PDSTestSuite) mustHavePDStestAccount(env environment) {
	// TODO: Use account name query filters
	accounts, resp, err := s.apiClient.AccountsApi.ApiAccountsGet(s.ctx).Execute()
	s.Require().NoError(err, "Error calling PDS API AccountsGet.")
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200")
	s.Require().NotEmpty(accounts, "PDS API must return at least one account.")

	var testPDSAccountID string
	for _, account := range accounts.GetData() {
		if account.GetName() == env.pdsAccountName {
			testPDSAccountID = account.GetId()
			break
		}
	}
	s.Require().NotEmpty(testPDSAccountID, "PDS account %s not found.", env.pdsAccountName)
	s.testPDSAccountID = testPDSAccountID
}

// mustHavePDStestTenant finds PDS tenant in Test PDS Account with name set in environment and stores its ID as "Test PDS Tenant".
func (s *PDSTestSuite) mustHavePDStestTenant(env environment) {
	// TODO: Use tenant name query filters
	tenants, resp, err := s.apiClient.TenantsApi.ApiAccountsIdTenantsGet(s.ctx, s.testPDSAccountID).Execute()
	s.Require().NoError(err, "Error calling PDS API AccountsIdTenantsGet.")
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200")
	s.Require().NotEmpty(tenants, "PDS API must return at least one tenant.")

	var testPDSTenantID string
	for _, tenant := range tenants.GetData() {
		if tenant.GetName() == env.pdsTenantName {
			testPDSTenantID = tenant.GetId()
			break
		}
	}
	s.Require().NotEmpty(testPDSTenantID, "PDS tenant %s not found.", env.pdsTenantName)
	s.testPDSTenantID = testPDSTenantID
}

// mustHavePDStestProject finds PDS project in Test PDS Tenant with name set in environment and stores its ID as "Test PDS Project".
func (s *PDSTestSuite) mustHavePDStestProject(env environment) {
	// TODO: Use project name query filters
	projects, resp, err := s.apiClient.ProjectsApi.ApiTenantsIdProjectsGet(s.ctx, s.testPDSTenantID).Execute()
	s.Require().NoError(err, "Error calling PDS API TenantsIdProjectsGet.")
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200")
	s.Require().NotEmpty(projects, "PDS API must return at least one project.")

	var testPDSProjectID string
	for _, project := range projects.GetData() {
		if project.GetName() == env.pdsProjectName {
			testPDSProjectID = project.GetId()
			break
		}
	}
	s.Require().NotEmpty(testPDSProjectID, "PDS project %s not found.", env.pdsProjectName)
	s.testPDSProjectID = testPDSProjectID
}

func (s *PDSTestSuite) mustHavePDStestDeploymentTarget(env environment) {
	var err error
	s.Require().Eventually(
		func() bool {
			s.testPDSDeploymentTargetID, err = getDeploymentTargetIDByName(s.ctx, s.apiClient, s.testPDSTenantID, env.pdsDeploymentTargetName)
			return err == nil
		},
		waiterDeploymentTargetNameExistsTimeout, waiterRetryInterval,
		"PDS deployment target %q does not exist: %v.", env.pdsDeploymentTargetName, err,
	)

	s.Require().Eventually(
		func() bool { return isDeploymentTargetHealthy(s.ctx, s.apiClient, s.testPDSDeploymentTargetID) },
		waiterDeploymentTargetStatusHealthyTimeout, waiterRetryInterval,
		"PDS deployment target %q is not healthy.", s.testPDSDeploymentTargetID,
	)
}

func (s *PDSTestSuite) mustDeletePDStestDeploymentTarget() {
	s.Require().Eventually(
		func() bool { return !isDeploymentTargetHealthy(s.ctx, s.apiClient, s.testPDSDeploymentTargetID) },
		waiterDeploymentTargetStatusUnhealthyTimeout, waiterRetryInterval,
		"PDS deployment target %s is still healthy.", s.testPDSDeploymentTargetID,
	)
	httpRes, err := s.apiClient.DeploymentTargetsApi.ApiDeploymentTargetsIdDelete(s.ctx, s.testPDSDeploymentTargetID).Execute()
	if err != nil {
		rawbody, parseErr := ioutil.ReadAll(httpRes.Body)
		s.Require().NoError(parseErr, "Parse api error")
		s.Require().NoError(err, "Error calling PDS API DeploymentTargetsIdDelete: %s", rawbody)
	}
	s.Require().Equal(204, httpRes.StatusCode, "PDS API must return HTTP 204")
}

func (s *PDSTestSuite) mustHavePDStestNamespace(env environment) {
	s.Require().Eventually(
		func() bool {
			var err error
			s.testPDSNamespaceID, err = getNamespaceIDByName(s.ctx, s.apiClient, s.testPDSDeploymentTargetID, env.pdsNamespaceName)
			return err == nil
		},
		waiterNamespaceExistsTimeout, waiterRetryInterval,
		"PDS Namespace %s does not exist.", env.pdsNamespaceName,
	)
}

// mustHavePDStestServiceAccount finds PDS Service account in Test PDS tenant with name set in environment and stores its ID as "Test PDS Service Account".
func (s *PDSTestSuite) mustHavePDStestServiceAccount(env environment) {
	// TODO: Use service account name query filters
	serviceAccounts, resp, err := s.apiClient.ServiceAccountsApi.ApiTenantsIdServiceAccountsGet(s.ctx, s.testPDSTenantID).Execute()
	s.Require().NoError(err, "Error calling PDS API TenantsIdServiceAccountsGet.")
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200")
	s.Require().NotEmpty(serviceAccounts, "PDS API must return at least one tenant.")

	var testPDSServiceAccountID string
	for _, serviceAccount := range serviceAccounts.GetData() {
		if serviceAccount.GetName() == env.pdsServiceAccountName {
			testPDSServiceAccountID = serviceAccount.GetId()
			break
		}
	}
	s.Require().NotEmpty(testPDSServiceAccountID, "PDS service account %s not found.", env.pdsServiceAccountName)
	s.testPDSServiceAccountID = testPDSServiceAccountID
}

// mustHavePDStestAgentToken gets "Test PDS Service Account" and stores its Token as "Test PDS Agent Token".
func (s *PDSTestSuite) mustHavePDStestAgentToken(env environment) {
	token, resp, err := s.apiClient.ServiceAccountsApi.ApiServiceAccountsIdTokenGet(s.ctx, s.testPDSServiceAccountID).Execute()
	s.Require().NoError(err, "Error calling PDS API ServiceAccountsIdTokenGet.")
	s.Require().Equal(200, resp.StatusCode, "PDS API must return HTTP 200.")

	s.testPDSAgentToken = token.GetToken()
}

func (s *PDSTestSuite) mustHaveAPIClient(env environment) {
	endpointUrl, err := url.Parse(env.controlPlaneAPI)
	s.Require().NoError(err, "Cannot parse control plane URL.")

	apiConf := pds.NewConfiguration()
	apiConf.Host = endpointUrl.Host
	apiConf.Scheme = endpointUrl.Scheme

	bearerToken, err := auth.GetBearerToken(s.ctx,
		env.secrets.tokenIssuerURL,
		env.secrets.issuerClientID,
		env.secrets.issuerClientSecret,
		env.secrets.pdsUsername,
		env.secrets.pdsPassword,
	)
	s.Require().NoError(err, "Cannot get bearer token.")
	s.ctx = context.WithValue(s.ctx,
		pds.ContextAPIKeys,
		map[string]pds.APIKey{
			"ApiKeyAuth": {Key: bearerToken, Prefix: "Bearer"},
		})

	s.apiClient = pds.NewAPIClient(apiConf)
}

func (s *PDSTestSuite) mustHaveTargetCluster(env environment) {
	tc, err := cluster.NewTargetCluster(env.targetKubeconfig)
	s.Require().NoError(err, "Cannot create target cluster.")
	s.targetCluster = tc
}

func (s *PDSTestSuite) mustInstallAgent(env environment) {
	provider, err := agent_installer.NewHelmProvider()
	s.Require().NoError(err, "Cannot create agent installer provider.")

	helmSelectorAgent14, err := agent_installer.NewSelectorHelmPDS18WithName(env.targetKubeconfig, s.testPDSTenantID, s.testPDSAgentToken, env.controlPlaneAPI, env.pdsDeploymentTargetName)
	s.Require().NoError(err, "Cannot create agent installer selector.")

	installer, err := provider.Installer(helmSelectorAgent14)
	s.Require().NoError(err, "Cannot get agent installer for version selector %s.", helmSelectorAgent14.ConstraintsString())

	err = installer.Install(s.ctx)
	s.Require().NoError(err, "Cannot install agent for version %s selector.", helmSelectorAgent14.ConstraintsString())
	s.pdsAgentInstallable = installer
}

func (s *PDSTestSuite) mustUninstallAgent(env environment) {
	err := s.targetCluster.DeleteCRDs(s.ctx)
	s.NoError(err, "Cannot delete CRDs.")
	err = s.pdsAgentInstallable.Uninstall(s.ctx)
	s.NoError(err, "Cannot uninstall agent.")
	err = s.targetCluster.DeleteClusterRoles(s.ctx)
	s.NoError(err, "Cannot delete cluster roles.")
	err = s.targetCluster.DeletePVCs(s.ctx, env.pdsNamespaceName)
	s.NoError(err, "Cannot delete PVCs.")
	err = s.targetCluster.DeleteStorageClasses(s.ctx)
	s.NoError(err, "Cannot delete storage classes.")
	err = s.targetCluster.DeleteReleasedPVs(s.ctx)
	s.NoError(err, "Cannot delete released PVs.")
	err = s.targetCluster.DeleteDetachedPXVolumes(s.ctx, env.pxNamespaceName)
	s.NoError(err, "Cannot delete detached PX volumes.")
	err = s.targetCluster.DeletePXCredentials(s.ctx, env.pxNamespaceName)
	s.NoError(err, "Cannot delete detached PX volumes.")

}

func (s *PDSTestSuite) mustLoadImageVersions() {
	imageVersions, err := getAllImageVersions(s.ctx, s.apiClient)
	s.Require().NoError(err, "Error while reading image versions.")
	s.Require().NotEmpty(imageVersions, "No image versions found.")
	s.imageVersionSpecList = imageVersions
}

func (s *PDSTestSuite) mustDeployDeploymentSpec(deployment ShortDeploymentSpec) string {
	image := findImageVersionForRecord(&deployment, s.imageVersionSpecList)
	s.Require().NotNil(image, "No image found for deployment %s.", deployment.ServiceName)

	deploymentID, err := createPDSDeployment(s.ctx, s.apiClient, &deployment, image, s.testPDSTenantID, s.testPDSDeploymentTargetID, s.testPDSProjectID, s.testPDSNamespaceID)
	s.Require().NoError(err, "Error while creating deployment %s.", deployment.ServiceName)
	s.Require().NotEmpty(deploymentID, "Deployment ID is empty.")

	return deploymentID
}

func (s *PDSTestSuite) mustUpdateDeployment(deploymentID string, spec *ShortDeploymentSpec) {
	req := pds.ControllersUpdateDeploymentRequest{}
	if spec.ImageVersionTag != "" || spec.ImageVersionBuild != "" {
		image := findImageVersionForRecord(spec, s.imageVersionSpecList)
		s.Require().NotNil(image, "Update deployment: no image found for %s version.", spec.ImageVersionTag)

		req.ImageId = &image.ImageID
	}
	if spec.NodeCount != 0 {
		nodeCount := int32(spec.NodeCount)
		req.NodeCount = &nodeCount
	}

	_, httpRes, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdPut(s.ctx, deploymentID).Body(req).Execute()
	if err != nil {
		rawbody, parseErr := ioutil.ReadAll(httpRes.Body)
		s.Require().NoError(parseErr, "Parse api error")
		s.Require().NoError(err, "Error calling PDS API DeploymentTargetsIdPut: %s", rawbody)
	}
	s.Require().NoError(err, "Update %s deployment.", deploymentID)
}

func (s *PDSTestSuite) mustEnsureDeploymentHealthy(deploymentID string) {
	s.Require().Eventually(
		func() bool {
			return isDeploymentHealthy(s.ctx, s.apiClient, deploymentID)
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Deployment %s is not healthy.", deploymentID,
	)
}

func (s *PDSTestSuite) mustEnsureStatefulSetReady(deploymentID string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	namespace := namespaceModel.GetName()
	s.Require().Eventually(
		func() bool {
			set, err := s.targetCluster.GetStatefulSet(s.ctx, namespace, deployment.GetClusterResourceName())
			if err != nil {
				return false
			}

			return *set.Spec.Replicas == set.Status.ReadyReplicas
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Deployment %s is not ready.", deploymentID,
	)
}

func (s *PDSTestSuite) mustEnsureStatefulSetReadyReplicas(deploymentID string, replicas int) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	namespace := namespaceModel.GetName()
	s.Require().Eventually(
		func() bool {
			set, err := s.targetCluster.GetStatefulSet(s.ctx, namespace, deployment.GetClusterResourceName())
			if err != nil {
				return false
			}

			return int(set.Status.ReadyReplicas) == replicas
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Deployment %s is expected to have %d ready replicas.", deploymentID, replicas,
	)
}

func (s *PDSTestSuite) mustEnsureStatefulSetImage(deploymentID, imageTag string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	dataService, _, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	s.Require().NoError(err)

	namespace := namespaceModel.GetName()
	s.Require().Eventually(
		func() bool {
			set, err := s.targetCluster.GetStatefulSet(s.ctx, namespace, deployment.GetClusterResourceName())
			if err != nil {
				return false
			}
			image, err := getDatabaseImage(dataService.GetName(), set)
			if err != nil {
				return false
			}
			return strings.Contains(image, imageTag)
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Statefulset %s is expected to have %s image tag.", deployment.GetClusterResourceName(), imageTag,
	)
}

func (s *PDSTestSuite) mustEnsureDeploymentInitialized(deploymentID string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	namespace := namespaceModel.GetName()
	clusterInitJobName := fmt.Sprintf("%s-cluster-init", deployment.GetClusterResourceName())
	nodeInitJobName := fmt.Sprintf("%s-node-init", deployment.GetClusterResourceName())

	s.Require().Eventually(
		func() bool {
			clusterInitJob, err := s.targetCluster.GetJob(s.ctx, namespace, clusterInitJobName)
			if err != nil {
				return false
			}
			if !isJobSucceeded(clusterInitJob) {
				return false
			}

			nodeInitJob, err := s.targetCluster.GetJob(s.ctx, namespace, nodeInitJobName)
			if err != nil {
				return false
			}
			return isJobSucceeded(nodeInitJob)
		},
		waiterDeploymentStatusHealthyTimeout, waiterRetryInterval,
		"Deployment %s is not ready.", deploymentID,
	)
}

func (s *PDSTestSuite) mustCreateBackup(deploymentID string) *pds.ModelsBackup {
	backupReq := pds.ControllersCreateDeploymentBackup{
		BackupLevel:    pointer.String("snapshot"),
		BackupTargetId: pointer.String(s.testPDSBackupTargetID),
		BackupType:     pointer.String("adhoc"),
	}
	backup, _, err := s.apiClient.BackupsApi.ApiDeploymentsIdBackupsPost(s.ctx, deploymentID).Body(backupReq).Execute()
	s.Require().NoError(err)

	return backup
}

func (s *PDSTestSuite) mustDeleteBackup(backupID string) {
	_, err := s.apiClient.BackupsApi.ApiBackupsIdDelete(s.ctx, backupID).Execute()
	s.Require().NoError(err)
}

func (s *PDSTestSuite) mustEnsureS3BackupCredentials(env environment) {
	tenantID := s.testPDSTenantID
	backupCredentialsName := fmt.Sprintf("%s-backupcredentials-s3", namePrefix)

	backupCredentialsResp, _, err := s.apiClient.BackupCredentialsApi.ApiTenantsIdBackupCredentialsGet(s.ctx, tenantID).Execute()
	s.Require().NoError(err)

	for _, backupCredentials := range backupCredentialsResp.Data {
		if backupCredentials.GetName() == backupCredentialsName {
			s.testPDSBackupCredentialsID = backupCredentials.GetId()
			break
		}
	}

	if s.testPDSBackupCredentialsID == "" {
		backupCredentials := pds.ControllersCreateBackupCredentialsRequest{
			Name: &backupCredentialsName,
			Credentials: &pds.ControllersCredentials{
				S3: &pds.ModelsS3Credentials{
					Endpoint:  pointer.String(env.backupTarget.credentials.s3.endpoint),
					AccessKey: pointer.String(env.backupTarget.credentials.s3.accessKey),
					SecretKey: pointer.String(env.backupTarget.credentials.s3.secretKey),
				},
			},
		}
		backupCreds, _, err := s.apiClient.BackupCredentialsApi.ApiTenantsIdBackupCredentialsPost(s.ctx, tenantID).Body(backupCredentials).Execute()
		s.Require().NoError(err)

		s.testPDSBackupCredentialsID = backupCreds.GetId()
	}
}

func (s *PDSTestSuite) mustEnsureS3BackupTarget(env environment) {
	tenantID := s.testPDSTenantID
	backupTargetName := fmt.Sprintf("%s-backuptarget-s3", namePrefix)

	backupTargets, _, err := s.apiClient.BackupTargetsApi.ApiTenantsIdBackupTargetsGet(s.ctx, tenantID).Execute()
	s.Require().NoError(err)

	for _, backupTarget := range backupTargets.Data {
		if backupTarget.GetName() == backupTargetName {
			s.testPDSBackupTargetID = backupTarget.GetId()
			break
		}
	}

	if s.testPDSBackupTargetID == "" {
		backupTarget := pds.ControllersCreateTenantBackupTarget{
			Name:                pointer.StringPtr(backupTargetName),
			BackupCredentialsId: pointer.StringPtr(s.testPDSBackupCredentialsID),
			Bucket:              pointer.String(env.backupTarget.bucket),
			Region:              pointer.String(env.backupTarget.region),
			Type:                pointer.String("s3"),
		}
		backupTargetModel, _, err := s.apiClient.BackupTargetsApi.ApiTenantsIdBackupTargetsPost(s.ctx, tenantID).Body(backupTarget).Execute()
		s.Require().NoError(err)

		s.testPDSBackupTargetID = backupTargetModel.GetId()
	}
}

func (s *PDSTestSuite) mustDeleteBackupCredentials() {
	if s.testPDSBackupCredentialsID != "" {
		_, err := s.apiClient.BackupCredentialsApi.ApiBackupCredentialsIdDelete(s.ctx, s.testPDSBackupCredentialsID).Execute()
		s.NoError(err)
	}
}

func (s *PDSTestSuite) mustDeleteBackupTarget() {
	if s.testPDSBackupTargetID != "" {
		_, err := s.apiClient.BackupTargetsApi.ApiBackupTargetsIdDelete(s.ctx, s.testPDSBackupTargetID).Execute()
		s.NoError(err)
	}
}

func (s *PDSTestSuite) mustEnsureBackupSuccessful(deploymentID, backupName string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespaceModel, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	namespace := namespaceModel.GetName()
	s.Require().Eventually(
		func() bool {
			pdsBackup, err := s.targetCluster.GetPDSBackup(s.ctx, namespace, backupName)
			if err != nil {
				return false
			}
			return isBackupSucceeded(pdsBackup)
		},
		waiterBackupStatusSucceededTimeout, waiterRetryInterval,
		"Backup %s for the %s deployment is not ready.", backupName, deploymentID,
	)
}

func (s *PDSTestSuite) mustRunBasicSmokeTest(deploymentID string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	dataService, _, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	s.Require().NoError(err)
	dataServiceType := dataService.GetName()

	switch dataServiceType {
	case dbPostgres:
		s.mustReadWriteData(deploymentID)
	default:
		s.mustRunLoadTestJob(deploymentID)
	}
}

func (s *PDSTestSuite) mustReadWriteData(deploymentID string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)

	namespace, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	dataService, _, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	s.Require().NoError(err)

	port, err := getDefaultDBPort(dataService.GetName())
	s.Require().NoError(err)

	podName := s.mustGetDeploymentPodName(deploymentID, deployment.GetClusterResourceName(), dataService.GetName(), namespace.GetName())
	pf, err := s.targetCluster.PortforwardPod(namespace.GetName(), podName, port)
	s.Require().NoError(err)

	dbPassword := s.mustGetDBPassword(namespace.GetName(), deployment.GetClusterResourceName())
	loadtest, err := s.mustGetLoadgenFor(dataService.GetName(), "localhost", strconv.Itoa(pf.Local), dbPassword)
	s.Require().NoError(err)

	err = loadtest.Run(s.ctx)
	s.Require().NoError(err)
}

func (s *PDSTestSuite) mustRunLoadTestJob(deploymentID string) {
	jobNamespace, jobName := s.mustCreateLoadTestJob(deploymentID)
	s.mustEnsureLoadTestJobSucceeded(jobNamespace, jobName)
	s.mustEnsureLoadTestJobLogsDoNotContain(jobNamespace, jobName, "ERROR|FATAL")
}

func (s *PDSTestSuite) mustCreateLoadTestJob(deploymentID string) (string, string) {
	deployment, _, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
	s.Require().NoError(err)
	deploymentName := deployment.GetClusterResourceName()

	namespace, _, err := s.apiClient.NamespacesApi.ApiNamespacesIdGet(s.ctx, *deployment.NamespaceId).Execute()
	s.Require().NoError(err)

	dataService, _, err := s.apiClient.DataServicesApi.ApiDataServicesIdGet(s.ctx, deployment.GetDataServiceId()).Execute()
	s.Require().NoError(err)
	dataServiceType := dataService.GetName()

	jobName := fmt.Sprintf("%s-loadtest", deployment.GetClusterResourceName())

	image, err := s.mustGetLoadTestJobImage(dataServiceType)
	s.Require().NoError(err)

	env := s.mustGetLoadTestJobEnv(deploymentName, namespace.GetName())

	job, err := s.targetCluster.CreateJob(s.ctx, namespace.GetName(), jobName, image, env)
	s.Require().NoError(err)

	return namespace.GetName(), job.GetName()
}

func (s *PDSTestSuite) mustEnsureLoadTestJobSucceeded(namespace, jobName string) {
	// 1. Wait for the job to finish.
	s.Require().Eventually(
		func() bool {
			job, err := s.targetCluster.GetJob(s.ctx, namespace, jobName)
			return err == nil && (job.Status.Succeeded > 0 || job.Status.Failed > 0)
		},
		waiterLoadTestJobFinishedTimeout, waiterRetryInterval,
		"Failed to wait for load test job %s to finish.", jobName,
	)

	// 2. Check the result.
	job, err := s.targetCluster.GetJob(s.ctx, namespace, jobName)
	s.Require().NoError(err)

	if job.Status.Failed > 0 {
		// Job failed.
		logs, err := s.targetCluster.GetJobLogs(s.T(), s.ctx, namespace, jobName, s.startTime)
		if err != nil {
			s.Require().Fail("Job '%s' failed.", jobName)
		} else {
			s.Require().Fail("Job '%s' failed. See job logs for more details:\n%s", jobName, logs)
		}
	}
	s.Require().True(job.Status.Succeeded > 0)
}

func (s *PDSTestSuite) mustEnsureLoadTestJobLogsDoNotContain(namespace, jobName, rePattern string) {
	logs, err := s.targetCluster.GetJobLogs(s.T(), s.ctx, namespace, jobName, s.startTime)
	s.Require().NoError(err)
	re := regexp.MustCompile(rePattern)
	s.Require().Nil(re.FindStringIndex(logs), "Job log '%s' contains pattern '%s':\n%s", jobName, rePattern, logs)
}

func (s *PDSTestSuite) mustGetLoadTestJobImage(dataServiceType string) (string, error) {
	switch dataServiceType {
	case dbCassandra:
		return "portworx/pds-loadtests:cassandra-0.0.3", nil
	default:
		return "", fmt.Errorf("loadtest job image not found for data service %s", dataServiceType)
	}
}

func (s *PDSTestSuite) mustGetLoadTestJobEnv(deploymentName, namespace string) []corev1.EnvVar {
	host := fmt.Sprintf("%s-%s", deploymentName, namespace)
	password := s.mustGetDBPassword(namespace, deploymentName)
	return []corev1.EnvVar{
		{
			Name:  "HOST",
			Value: host,
		}, {
			Name:  "PASSWORD",
			Value: password,
		}, {
			Name:  "ITERATIONS",
			Value: "1",
		}, {
			Name:  "FAIL_ON_ERROR",
			Value: "true",
		}}
}

func (s *PDSTestSuite) mustRemoveDeployment(deploymentID string) {
	_, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdDelete(s.ctx, deploymentID).Execute()
	s.Require().NoError(err, "Error while removing deployment %s.", deploymentID)
}

func (s *PDSTestSuite) mustEnsureDeploymentRemoved(deploymentID string) {
	s.Require().Eventually(
		func() bool {
			_, httpResp, err := s.apiClient.DeploymentsApi.ApiDeploymentsIdGet(s.ctx, deploymentID).Execute()
			return httpResp != nil && httpResp.StatusCode == 404 && err != nil
		},
		waiterDeploymentStatusRemovedTimeout, waiterRetryInterval,
		"Deployment %s is not removed.", deploymentID,
	)
}

func (s *PDSTestSuite) mustHaveTargetClusterNamespaces(env environment) {
	nss := []string{env.pdsNamespaceName}
	s.targetCluster.EnsureNamespaces(s.T(), s.ctx, nss)
}

func (s *PDSTestSuite) mustGetDBPassword(namespace, deploymentName string) string {
	secretName := fmt.Sprintf("%s-creds", deploymentName)
	secret, err := s.targetCluster.GetSecret(s.ctx, namespace, secretName)
	s.Require().NoError(err)

	return string(secret.Data["password"])
}

func (s *PDSTestSuite) mustGetLoadgenFor(dataServiceType, host, port, dbPassword string) (loadgen.Interface, error) {
	switch dataServiceType {
	case dbPostgres:
		return s.newPostgresLoadgen(host, port, "pds", dbPassword), nil
	}

	return nil, fmt.Errorf("loadgen for the %s database is not found", dataServiceType)
}

func (s *PDSTestSuite) newPostgresLoadgen(host, port, dbUser, dbPassword string) loadgen.Interface {
	postgresLoader, err := postgresql.New(postgresql.Config{
		User:     dbUser,
		Password: dbPassword,
		Host:     host,
		Port:     port,
		Count:    5,
		Logger:   newTestLogger(s.T()),
	})
	s.Require().NoError(err)

	return postgresLoader
}

func (s *PDSTestSuite) mustGetDeploymentPodName(deploymentID, deploymentName, deploymentType, namespace string) string {
	switch deploymentType {
	case dbPostgres:
		// For the port-forward tunnel needs the 'master' pod otherwise we'll get the 'read-only' error.
		selector := map[string]string{
			"pds/deployment-id": deploymentID,
			"role":              "master",
		}
		podList, err := s.targetCluster.ListPods(s.ctx, namespace, selector)
		s.Require().NoError(err)
		s.Require().NotEmpty(podList.Items, "get pods for the label selector: %s", labels.FormatLabels(selector))
		return podList.Items[0].Name
	default:
		return getDeploymentPodName(deploymentName)
	}
}

func getDefaultDBPort(deploymentType string) (int, error) {
	switch deploymentType {
	case dbPostgres:
		return 5432, nil
	}
	return -1, fmt.Errorf("has no default port for the %s database", deploymentType)
}

func getDatabaseImage(deploymentType string, set *appsv1.StatefulSet) (string, error) {
	var containerName string
	switch deploymentType {
	case dbPostgres:
		containerName = "postgresql"
	case dbCassandra:
		containerName = "cassandra"
	default:
		return "", fmt.Errorf("unknown database type: %s", deploymentType)
	}

	for _, container := range set.Spec.Template.Spec.Containers {
		if container.Name != containerName {
			continue
		}

		return container.Image, nil
	}

	return "", fmt.Errorf("database type: %s: container %q is not found", deploymentType, containerName)
}
