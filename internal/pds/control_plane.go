package pds

import (
	"context"
	"fmt"
	"net/url"
	"time"

	pdsApi "github.com/portworx/pds-api-go-client/pds/v1alpha1"

	"github.com/portworx/pds-integration-test/internal/api"
	"github.com/portworx/pds-integration-test/internal/auth"
)

const DefaultActor = "default-pds-actor"

type LoginCredentials struct {
	TokenIssuerURL     string
	IssuerClientID     string
	IssuerClientSecret string
	Username           string
	Password           string
	BearerToken        string
}

type ActorContext struct {
	AccountID string
	TenantID  string
	ProjectID string
	AuthCtx   context.Context
}

type ControlPlane struct {
	ApiClient    *pdsApi.APIClient
	actorContext map[string]ActorContext
}

func NewControlPlane(apiClient *pdsApi.APIClient, actorContext ActorContext) *ControlPlane {
	cp := ControlPlane{
		ApiClient: apiClient,
		actorContext: map[string]ActorContext{
			DefaultActor: actorContext,
		},
	}

	return &cp
}

func CreateActorContextUsingApiClient(credentials LoginCredentials, accountName string, tenantName string,
	projectName string, apiClient *pdsApi.APIClient) (*ActorContext, error) {
	cp := ControlPlane{
		ApiClient: apiClient,
	}
	return cp.CreateActorContext(credentials, accountName, tenantName, projectName)
}

// CreateActorContext creates ActorContext struct, that can be used to authorize apiClient API calls. If
// accountName, tenantName and projectName are set, their IDs will be filled in the final struct (that's useful for
// working with PDS API resources).
func (c *ControlPlane) CreateActorContext(credentials LoginCredentials,
	accountName string, tenantName string, projectName string) (*ActorContext, error) {
	if tenantName != "" && accountName == "" {
		return nil, fmt.Errorf("tenant name can't be used with empty account name")
	}
	if projectName != "" && (accountName == "" || tenantName == "") {
		return nil, fmt.Errorf("project name can be used only when account name and tenant name is provided too")
	}

	authContext, err := CreateAuthContext(context.Background(), credentials)
	if err != nil {
		return nil, fmt.Errorf("could not authenticate: %w", err)
	}
	actorContext := ActorContext{AuthCtx: authContext}
	if accountName == "" {
		return &actorContext, nil
	}

	account, err := c.GetAccount(accountName, actorContext)
	if err != nil {
		return nil, fmt.Errorf("could not find PDS account: %w", err)
	}
	actorContext.AccountID = account.GetId()
	if tenantName == "" {
		return &actorContext, nil
	}

	tenant, err := c.GetTenant(tenantName, actorContext)
	if err != nil {
		return nil, fmt.Errorf("could not find PDS tenant: %w", err)
	}
	actorContext.TenantID = tenant.GetId()
	if projectName == "" {
		return &actorContext, nil
	}

	project, err := c.GetProject(projectName, actorContext)
	if err != nil {
		return nil, fmt.Errorf("could not find PDS project: %w", err)
	}
	actorContext.ProjectID = project.GetId()
	return &actorContext, nil
}

func (c *ControlPlane) CreateUserAPIKey(expiresAt time.Time, name string,
	actor ActorContext) (*pdsApi.ModelsUserAPIKey, error) {

	expirationDate := expiresAt.Format(time.RFC3339)
	requestBody := pdsApi.RequestsCreateUserAPIKeyRequest{
		ExpiresAt: &expirationDate,
		Name:      &name,
	}
	userApiKey, response, err := c.ApiClient.UserAPIKeyApi.ApiUserApiKeyPost(actor.AuthCtx).Body(requestBody).Execute()
	err = api.ExtractErrorDetails(response, err)
	if err != nil {
		return nil, fmt.Errorf("could not create user API key: %w", err)
	}
	return userApiKey, nil
}

// GetDefaultActor - default actor is the user on behalf which the API calls will be performed
// (if not explicitly asked to use a different user)
func (c *ControlPlane) GetDefaultActor() ActorContext {
	return c.actorContext[DefaultActor]
}

func CreateAuthContext(
	ctx context.Context, credentials LoginCredentials) (context.Context, error) {
	bearerToken := credentials.BearerToken
	if bearerToken == "" {
		var err error
		bearerToken, err = auth.GetBearerToken(ctx,
			credentials.TokenIssuerURL,
			credentials.IssuerClientID,
			credentials.IssuerClientSecret,
			credentials.Username,
			credentials.Password,
		)
		if err != nil {
			return nil, fmt.Errorf("could not get bearer token: %w", err)
		}
	}

	ctx = context.WithValue(ctx,
		pdsApi.ContextAPIKeys,
		map[string]pdsApi.APIKey{
			"ApiKeyAuth": {Key: bearerToken, Prefix: "Bearer"},
		})
	return ctx, nil
}

func CreateAPIClient(controlPlaneApiUrl string) (*pdsApi.APIClient, error) {
	endpointUrl, err := url.Parse(controlPlaneApiUrl)
	if err != nil {
		return nil, fmt.Errorf("could not parse the Control Plane API URL: %w", err)
	}

	apiConf := pdsApi.NewConfiguration()
	apiConf.Host = endpointUrl.Host
	apiConf.Scheme = endpointUrl.Scheme
	return pdsApi.NewAPIClient(apiConf), nil
}

func (c *ControlPlane) GetAccount(
	accountName string, actor ActorContext) (*pdsApi.ModelsAccount, error) {

	accounts, response, err := c.ApiClient.AccountsApi.ApiAccountsGet(actor.AuthCtx).Execute()
	if err := api.ExtractErrorDetails(response, err); err != nil {
		return nil, fmt.Errorf("could not get PDS Account name: %w", err)
	}

	for _, account := range accounts.GetData() {
		if account.GetName() == accountName {
			return &account, nil
		}
	}
	return nil, fmt.Errorf("account %q was not found", accountName)
}

func (c *ControlPlane) GetTenant(tenantName string, actor ActorContext) (*pdsApi.ModelsTenant, error) {

	tenants, response, err := c.ApiClient.TenantsApi.ApiAccountsIdTenantsGet(actor.AuthCtx, actor.AccountID).Execute()
	if err := api.ExtractErrorDetails(response, err); err != nil {
		return nil, fmt.Errorf("could not get PDS Tenant name: %w", err)
	}

	for _, tenant := range tenants.GetData() {
		if tenant.GetName() == tenantName {
			return &tenant, nil
		}
	}
	return nil, fmt.Errorf("tenant %q was not found in account with ID %q", tenantName, actor.AccountID)
}

func (c *ControlPlane) GetProject(projectName string, actor ActorContext) (*pdsApi.ModelsProject, error) {

	projects, response, err := c.ApiClient.ProjectsApi.ApiTenantsIdProjectsGet(actor.AuthCtx, actor.TenantID).Execute()
	if err := api.ExtractErrorDetails(response, err); err != nil {
		return nil, fmt.Errorf("could not get PDS Project name: %w", err)
	}

	for _, project := range projects.GetData() {
		if project.GetName() == projectName {
			return &project, nil
		}
	}
	return nil, fmt.Errorf("project %q was not found under tenant with ID %q", projectName, actor.TenantID)
}
