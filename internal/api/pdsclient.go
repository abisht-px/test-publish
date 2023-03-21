package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	pdsApi "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/auth"
)

const (
	pdsDeploymentTargetHealthStateHealthy = "healthy"
)

// PDSClient is a wrapper around the PDS OpenAPI client with additional helper logic.
type PDSClient struct {
	*pdsApi.APIClient
	URL string
}

type LoginCredentials struct {
	TokenIssuerURL     string
	IssuerClientID     string
	IssuerClientSecret string
	Username           string
	Password           string
	BearerToken        string
}

func NewPDSClient(ctx context.Context, apiURL string, credentials LoginCredentials) (*PDSClient, error) {
	endpointUrl, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("parsing API URL %q: %w", apiURL, err)
	}

	apiConf := pdsApi.NewConfiguration()
	apiConf.Host = endpointUrl.Host
	apiConf.Scheme = endpointUrl.Scheme
	httpClient, err := createAuthenticatedHTTPClient(ctx, credentials)
	if err != nil {
		return nil, fmt.Errorf("creating authenticated client: %w", err)
	}
	apiConf.HTTPClient = httpClient

	return &PDSClient{
		APIClient: pdsApi.NewAPIClient(apiConf),
		URL:       apiURL,
	}, nil
}

func createAuthenticatedHTTPClient(ctx context.Context, credentials LoginCredentials) (*http.Client, error) {
	var httpClient *http.Client
	bearerToken := credentials.BearerToken
	if bearerToken == "" {
		var err error
		client, err := auth.GetAuthenticatedClientByPassword(ctx,
			credentials.TokenIssuerURL,
			credentials.IssuerClientID,
			credentials.IssuerClientSecret,
			credentials.Username,
			credentials.Password,
		)
		if err != nil {
			return nil, fmt.Errorf("creating authenticated http client: %w", err)
		}

		httpClient = client
	} else {
		httpClient = auth.GetAuthenticatedClientByToken(ctx, bearerToken)
	}
	return httpClient, nil
}

func (c *PDSClient) GetAccount(accountName string) (*pdsApi.ModelsAccount, error) {
	accounts, response, err := c.AccountsApi.ApiAccountsGet(context.Background()).Execute()
	if err := ExtractErrorDetails(response, err); err != nil {
		return nil, fmt.Errorf("could not get PDS Account name: %w", err)
	}

	for _, account := range accounts.GetData() {
		if account.GetName() == accountName {
			return &account, nil
		}
	}
	return nil, fmt.Errorf("account %q was not found", accountName)
}

func (c *PDSClient) GetTenant(accountID, tenantName string) (*pdsApi.ModelsTenant, error) {
	tenants, response, err := c.TenantsApi.ApiAccountsIdTenantsGet(context.Background(), accountID).Execute()
	if err := ExtractErrorDetails(response, err); err != nil {
		return nil, fmt.Errorf("could not get PDS Tenant name: %w", err)
	}

	for _, tenant := range tenants.GetData() {
		if tenant.GetName() == tenantName {
			return &tenant, nil
		}
	}
	return nil, fmt.Errorf("tenant %q was not found in account with ID %q", tenantName, accountID)
}

func (c *PDSClient) GetProject(tenantID, projectName string) (*pdsApi.ModelsProject, error) {
	projects, response, err := c.ProjectsApi.ApiTenantsIdProjectsGet(context.Background(), tenantID).Execute()
	if err := ExtractErrorDetails(response, err); err != nil {
		return nil, fmt.Errorf("could not get PDS Project name: %w", err)
	}

	for _, project := range projects.GetData() {
		if project.GetName() == projectName {
			return &project, nil
		}
	}
	return nil, fmt.Errorf("project %q was not found under tenant with ID %q", projectName, tenantID)
}

func (c *PDSClient) CreateUserAPIKey(expiresAt time.Time, name string) (*pdsApi.ModelsUserAPIKey, error) {
	expirationDate := expiresAt.Format(time.RFC3339)
	requestBody := pdsApi.RequestsCreateUserAPIKeyRequest{
		ExpiresAt: &expirationDate,
		Name:      &name,
	}
	userApiKey, response, err := c.UserAPIKeyApi.ApiUserApiKeyPost(context.Background()).Body(requestBody).Execute()
	err = ExtractErrorDetails(response, err)
	if err != nil {
		return nil, fmt.Errorf("could not create user API key: %w", err)
	}
	return userApiKey, nil
}

func (c *PDSClient) CheckDeploymentTargetHealth(ctx context.Context, deploymentTargetID string) error {
	target, resp, err := c.DeploymentTargetsApi.ApiDeploymentTargetsIdGet(ctx, deploymentTargetID).Execute()
	if err != nil {
		return ExtractErrorDetails(resp, err)
	}
	if target.GetStatus() != pdsDeploymentTargetHealthStateHealthy {
		return fmt.Errorf("deployment target not healthy: got %q, want %q", target.GetStatus(), pdsDeploymentTargetHealthStateHealthy)
	}
	return nil
}

func (c *PDSClient) GetDeploymentTargetIDByName(ctx context.Context, tenantID, deploymentTargetName string) (string, error) {
	targets, resp, err := c.DeploymentTargetsApi.ApiTenantsIdDeploymentTargetsGet(ctx, tenantID).Execute()
	if err = ExtractErrorDetails(resp, err); err != nil {
		return "", fmt.Errorf("getting deployment targets for tenant %s: %w", tenantID, err)
	}
	for _, target := range targets.GetData() {
		if target.GetName() == deploymentTargetName {
			return target.GetId(), nil
		}
	}
	return "", fmt.Errorf("deployment target %s not found", deploymentTargetName)
}

func (c *PDSClient) GetNamespaceByName(ctx context.Context, deploymentTargetID, name string) (*pdsApi.ModelsNamespace, error) {
	namespaces, resp, err := c.NamespacesApi.ApiDeploymentTargetsIdNamespacesGet(ctx, deploymentTargetID).Execute()
	if err = ExtractErrorDetails(resp, err); err != nil {
		return nil, fmt.Errorf("getting namespace %s: %w", name, err)
	}
	for _, namespace := range namespaces.GetData() {
		if namespace.GetName() == name {
			return &namespace, nil
		}
	}
	return nil, nil
}

func (c *PDSClient) GetAllImageVersions(ctx context.Context) ([]PDSImageReferenceSpec, error) {
	var records []PDSImageReferenceSpec

	dataServices, resp, err := c.DataServicesApi.ApiDataServicesGet(ctx).Execute()
	if err = ExtractErrorDetails(resp, err); err != nil {
		return nil, fmt.Errorf("fetching all data services: %w", err)
	}

	dataServicesByID := make(map[string]pdsApi.ModelsDataService)
	for i := range dataServices.GetData() {
		dataService := dataServices.GetData()[i]
		dataServicesByID[dataService.GetId()] = dataService
	}

	images, resp, err := c.ImagesApi.ApiImagesGet(ctx).Latest(true).SortBy("-created_at").Limit("1000").Execute()
	if err = ExtractErrorDetails(resp, err); err != nil {
		return nil, fmt.Errorf("fetching all images: %w", err)
	}

	for _, image := range images.GetData() {
		dataService := dataServicesByID[image.GetDataServiceId()]
		record := PDSImageReferenceSpec{
			DataServiceName:   dataService.GetName(),
			DataServiceID:     dataService.GetId(),
			VersionID:         image.GetVersionId(),
			ImageVersionBuild: image.GetBuild(),
			ImageVersionTag:   image.GetTag(),
			ImageID:           image.GetId(),
		}
		records = append(records, record)
	}

	return records, nil
}

func (c *PDSClient) GetResourceSettingsTemplateByName(ctx context.Context, tenantID, templateName, dataServiceID string) (*pdsApi.ModelsResourceSettingsTemplate, error) {
	resources, resp, err := c.ResourceSettingsTemplatesApi.ApiTenantsIdResourceSettingsTemplatesGet(ctx, tenantID).Name(templateName).Execute()
	if err = ExtractErrorDetails(resp, err); err != nil {
		return nil, err
	}
	for _, r := range resources.GetData() {
		if r.GetDataServiceId() == dataServiceID {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("resource settings template %s not found", templateName)
}

func (c *PDSClient) GetAppConfigTemplateByName(ctx context.Context, tenantID, templateName, dataServiceID string) (*pdsApi.ModelsApplicationConfigurationTemplate, error) {
	appConfigurations, resp, err := c.ApplicationConfigurationTemplatesApi.ApiTenantsIdApplicationConfigurationTemplatesGet(ctx, tenantID).Name(templateName).Execute()
	if err = ExtractErrorDetails(resp, err); err != nil {
		return nil, err
	}
	for _, c := range appConfigurations.GetData() {
		if c.GetDataServiceId() == dataServiceID {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("application configuration template %s not found", templateName)
}

func (c *PDSClient) CreateDeployment(ctx context.Context, deployment *ShortDeploymentSpec, image *PDSImageReferenceSpec, tenantID, deploymentTargetID, projectID, namespaceID string) (string, error) {
	resource, err := c.GetResourceSettingsTemplateByName(ctx, tenantID, deployment.ResourceSettingsTemplateName, image.DataServiceID)
	if err != nil {
		return "", fmt.Errorf("getting resource settings template %s for tenant %s: %w", deployment.ResourceSettingsTemplateName, tenantID, err)
	}

	appConfig, err := c.GetAppConfigTemplateByName(ctx, tenantID, deployment.AppConfigTemplateName, image.DataServiceID)
	if err != nil {
		return "", fmt.Errorf("getting application configuration template %s for tenant %s: %w", deployment.AppConfigTemplateName, tenantID, err)
	}

	storages, resp, err := c.StorageOptionsTemplatesApi.ApiTenantsIdStorageOptionsTemplatesGet(ctx, tenantID).Name(deployment.StorageOptionName).Execute()
	if err = ExtractErrorDetails(resp, err); err != nil {
		return "", fmt.Errorf("getting storage option template %s for tenant %s: %w", deployment.StorageOptionName, tenantID, err)
	}

	if len(storages.GetData()) == 0 {
		return "", fmt.Errorf("storage option template %s not found", deployment.StorageOptionName)
	}
	if len(storages.GetData()) != 1 {
		return "", fmt.Errorf("more than one storage option template found")
	}
	storage := storages.GetData()[0]

	var backupPolicy *pdsApi.ModelsBackupPolicy
	if len(deployment.BackupPolicyname) > 0 {
		backupPolicies, resp, err := c.BackupPoliciesApi.ApiTenantsIdBackupPoliciesGet(ctx, tenantID).Name(deployment.BackupPolicyname).Execute()
		if err = ExtractErrorDetails(resp, err); err != nil {
			return "", fmt.Errorf("getting backup policies for tenant %s: %w", tenantID, err)
		}
		if len(backupPolicies.GetData()) == 0 {
			return "", fmt.Errorf("backup policy %s not found", deployment.BackupPolicyname)
		}
		if len(backupPolicies.GetData()) != 1 {
			return "", fmt.Errorf("more than one backup policy found")
		}
		backupPolicy = &backupPolicies.GetData()[0]
	}

	dns, resp, err := c.TenantsApi.ApiTenantsIdDnsDetailsGet(ctx, tenantID).Execute()
	if err = ExtractErrorDetails(resp, err); err != nil {
		return "", fmt.Errorf("getting DNS details for tenant %s: %w", tenantID, err)
	}

	pdsDeployment := pdsApi.NewControllersCreateProjectDeployment()
	pdsDeployment.SetApplicationConfigurationTemplateId(appConfig.GetId())
	pdsDeployment.SetDeploymentTargetId(deploymentTargetID)
	pdsDeployment.SetDnsZone(dns.GetDnsZone())
	pdsDeployment.SetImageId(image.ImageID)
	pdsDeployment.SetName(deployment.NamePrefix)
	pdsDeployment.SetNamespaceId(namespaceID)
	pdsDeployment.SetNodeCount(int32(deployment.NodeCount))
	pdsDeployment.SetResourceSettingsTemplateId(resource.GetId())
	if backupPolicy != nil {
		pdsDeployment.ScheduledBackup.SetBackupPolicyId(backupPolicy.GetId())
	}
	pdsDeployment.SetServiceType(deployment.ServiceType)
	pdsDeployment.SetStorageOptionsTemplateId(storage.GetId())

	res, httpRes, err := c.DeploymentsApi.ApiProjectsIdDeploymentsPost(ctx, projectID).Body(*pdsDeployment).Execute()
	if err = ExtractErrorDetails(httpRes, err); err != nil {
		return "", fmt.Errorf("deploying %s under project %s: %w", *pdsDeployment.Name, projectID, err)
	}

	return res.GetId(), nil
}
