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
