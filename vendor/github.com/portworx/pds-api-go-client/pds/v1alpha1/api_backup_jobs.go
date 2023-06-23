/*
PDS API

Portworx Data Services API Server

API version: 1.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package pds

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Linger please
var (
	_ context.Context
)

// BackupJobsApiService BackupJobsApi service
type BackupJobsApiService service

type ApiApiBackupJobsIdDeleteRequest struct {
	ctx context.Context
	ApiService *BackupJobsApiService
	id string
}


func (r ApiApiBackupJobsIdDeleteRequest) Execute() (*http.Response, error) {
	return r.ApiService.ApiBackupJobsIdDeleteExecute(r)
}

/*
ApiBackupJobsIdDelete Delete BackupJob

Removes a single BackupJob

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id BackupJob ID (must be valid UUID)
 @return ApiApiBackupJobsIdDeleteRequest
*/
func (a *BackupJobsApiService) ApiBackupJobsIdDelete(ctx context.Context, id string) ApiApiBackupJobsIdDeleteRequest {
	return ApiApiBackupJobsIdDeleteRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
func (a *BackupJobsApiService) ApiBackupJobsIdDeleteExecute(r ApiApiBackupJobsIdDeleteRequest) (*http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodDelete
		localVarPostBody     interface{}
		formFiles            []formFile
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "BackupJobsApiService.ApiBackupJobsIdDelete")
	if err != nil {
		return nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/backup-jobs/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarHTTPResponse, newErr
	}

	return localVarHTTPResponse, nil
}

type ApiApiBackupJobsIdGetRequest struct {
	ctx context.Context
	ApiService *BackupJobsApiService
	id string
}


func (r ApiApiBackupJobsIdGetRequest) Execute() (*ModelsBackupJob, *http.Response, error) {
	return r.ApiService.ApiBackupJobsIdGetExecute(r)
}

/*
ApiBackupJobsIdGet Get BackupJob

Fetches a BackupJob

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id BackupJob ID (must be valid UUID)
 @return ApiApiBackupJobsIdGetRequest
*/
func (a *BackupJobsApiService) ApiBackupJobsIdGet(ctx context.Context, id string) ApiApiBackupJobsIdGetRequest {
	return ApiApiBackupJobsIdGetRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return ModelsBackupJob
func (a *BackupJobsApiService) ApiBackupJobsIdGetExecute(r ApiApiBackupJobsIdGetRequest) (*ModelsBackupJob, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ModelsBackupJob
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "BackupJobsApiService.ApiBackupJobsIdGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/backup-jobs/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiApiBackupJobsIdPutRequest struct {
	ctx context.Context
	ApiService *BackupJobsApiService
	id string
	body *RequestsPutBackupJobRequest
}

// Request body containing backup job details
func (r ApiApiBackupJobsIdPutRequest) Body(body RequestsPutBackupJobRequest) ApiApiBackupJobsIdPutRequest {
	r.body = &body
	return r
}

func (r ApiApiBackupJobsIdPutRequest) Execute() (*ModelsBackupJob, *http.Response, error) {
	return r.ApiService.ApiBackupJobsIdPutExecute(r)
}

/*
ApiBackupJobsIdPut Upsert BackupJob

Updates a BackupJob if ID exists, Creates new BackupJob if not

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id BackupJob ID (must be UUID)
 @return ApiApiBackupJobsIdPutRequest
*/
func (a *BackupJobsApiService) ApiBackupJobsIdPut(ctx context.Context, id string) ApiApiBackupJobsIdPutRequest {
	return ApiApiBackupJobsIdPutRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return ModelsBackupJob
func (a *BackupJobsApiService) ApiBackupJobsIdPutExecute(r ApiApiBackupJobsIdPutRequest) (*ModelsBackupJob, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodPut
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ModelsBackupJob
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "BackupJobsApiService.ApiBackupJobsIdPut")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/backup-jobs/{id}"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}
	if r.body == nil {
		return localVarReturnValue, nil, reportError("body is required and must be specified")
	}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{"application/json"}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	// body params
	localVarPostBody = r.body
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiApiBackupsIdJobsGetRequest struct {
	ctx context.Context
	ApiService *BackupJobsApiService
	id string
}


func (r ApiApiBackupsIdJobsGetRequest) Execute() (*ControllersListBackupJobsStatusResponse, *http.Response, error) {
	return r.ApiService.ApiBackupsIdJobsGetExecute(r)
}

/*
ApiBackupsIdJobsGet List Backup's Jobs

Retrieves a list of BackupJobs associated to this Backup

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id Backup ID (must be valid UUID)
 @return ApiApiBackupsIdJobsGetRequest
*/
func (a *BackupJobsApiService) ApiBackupsIdJobsGet(ctx context.Context, id string) ApiApiBackupsIdJobsGetRequest {
	return ApiApiBackupsIdJobsGetRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return ControllersListBackupJobsStatusResponse
func (a *BackupJobsApiService) ApiBackupsIdJobsGetExecute(r ApiApiBackupsIdJobsGetRequest) (*ControllersListBackupJobsStatusResponse, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ControllersListBackupJobsStatusResponse
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "BackupJobsApiService.ApiBackupsIdJobsGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/backups/{id}/jobs"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}

type ApiApiProjectsIdBackupJobsGetRequest struct {
	ctx context.Context
	ApiService *BackupJobsApiService
	id string
	sortBy *string
	limit *string
	continuation *string
	id2 *string
	deploymentId *string
	deploymentTargetId *string
	backupId *string
	name *string
	namespaceId *string
}

// A given BackupJob attribute to sort results by (one of: id, name, created_at)
func (r ApiApiProjectsIdBackupJobsGetRequest) SortBy(sortBy string) ApiApiProjectsIdBackupJobsGetRequest {
	r.sortBy = &sortBy
	return r
}
// Maximum number of rows to return (could be less)
func (r ApiApiProjectsIdBackupJobsGetRequest) Limit(limit string) ApiApiProjectsIdBackupJobsGetRequest {
	r.limit = &limit
	return r
}
// Use a token returned by a previous query to continue listing with the next batch of rows
func (r ApiApiProjectsIdBackupJobsGetRequest) Continuation(continuation string) ApiApiProjectsIdBackupJobsGetRequest {
	r.continuation = &continuation
	return r
}
// Filter results by BackupJob id
func (r ApiApiProjectsIdBackupJobsGetRequest) Id2(id2 string) ApiApiProjectsIdBackupJobsGetRequest {
	r.id2 = &id2
	return r
}
// Filter results by BackupJob deployment_id
func (r ApiApiProjectsIdBackupJobsGetRequest) DeploymentId(deploymentId string) ApiApiProjectsIdBackupJobsGetRequest {
	r.deploymentId = &deploymentId
	return r
}
// Filter results by BackupJob deployment_target_id
func (r ApiApiProjectsIdBackupJobsGetRequest) DeploymentTargetId(deploymentTargetId string) ApiApiProjectsIdBackupJobsGetRequest {
	r.deploymentTargetId = &deploymentTargetId
	return r
}
// Filter results by BackupJob backup_id
func (r ApiApiProjectsIdBackupJobsGetRequest) BackupId(backupId string) ApiApiProjectsIdBackupJobsGetRequest {
	r.backupId = &backupId
	return r
}
// Filter results by BackupJob name
func (r ApiApiProjectsIdBackupJobsGetRequest) Name(name string) ApiApiProjectsIdBackupJobsGetRequest {
	r.name = &name
	return r
}
// Filter results by BackupJob namespace_id
func (r ApiApiProjectsIdBackupJobsGetRequest) NamespaceId(namespaceId string) ApiApiProjectsIdBackupJobsGetRequest {
	r.namespaceId = &namespaceId
	return r
}

func (r ApiApiProjectsIdBackupJobsGetRequest) Execute() (*ModelsPaginatedResultModelsBackupJob, *http.Response, error) {
	return r.ApiService.ApiProjectsIdBackupJobsGetExecute(r)
}

/*
ApiProjectsIdBackupJobsGet List Project's BackupJobs

Lists the BackupJobs that belonging to the Project.

 @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 @param id Project ID (must be valid UUID)
 @return ApiApiProjectsIdBackupJobsGetRequest
*/
func (a *BackupJobsApiService) ApiProjectsIdBackupJobsGet(ctx context.Context, id string) ApiApiProjectsIdBackupJobsGetRequest {
	return ApiApiProjectsIdBackupJobsGetRequest{
		ApiService: a,
		ctx: ctx,
		id: id,
	}
}

// Execute executes the request
//  @return ModelsPaginatedResultModelsBackupJob
func (a *BackupJobsApiService) ApiProjectsIdBackupJobsGetExecute(r ApiApiProjectsIdBackupJobsGetRequest) (*ModelsPaginatedResultModelsBackupJob, *http.Response, error) {
	var (
		localVarHTTPMethod   = http.MethodGet
		localVarPostBody     interface{}
		formFiles            []formFile
		localVarReturnValue  *ModelsPaginatedResultModelsBackupJob
	)

	localBasePath, err := a.client.cfg.ServerURLWithContext(r.ctx, "BackupJobsApiService.ApiProjectsIdBackupJobsGet")
	if err != nil {
		return localVarReturnValue, nil, &GenericOpenAPIError{error: err.Error()}
	}

	localVarPath := localBasePath + "/api/projects/{id}/backup-jobs"
	localVarPath = strings.Replace(localVarPath, "{"+"id"+"}", url.PathEscape(parameterToString(r.id, "")), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if r.sortBy != nil {
		localVarQueryParams.Add("sort_by", parameterToString(*r.sortBy, ""))
	}
	if r.limit != nil {
		localVarQueryParams.Add("limit", parameterToString(*r.limit, ""))
	}
	if r.continuation != nil {
		localVarQueryParams.Add("continuation", parameterToString(*r.continuation, ""))
	}
	if r.id2 != nil {
		localVarQueryParams.Add("id", parameterToString(*r.id2, ""))
	}
	if r.deploymentId != nil {
		localVarQueryParams.Add("deployment_id", parameterToString(*r.deploymentId, ""))
	}
	if r.deploymentTargetId != nil {
		localVarQueryParams.Add("deployment_target_id", parameterToString(*r.deploymentTargetId, ""))
	}
	if r.backupId != nil {
		localVarQueryParams.Add("backup_id", parameterToString(*r.backupId, ""))
	}
	if r.name != nil {
		localVarQueryParams.Add("name", parameterToString(*r.name, ""))
	}
	if r.namespaceId != nil {
		localVarQueryParams.Add("namespace_id", parameterToString(*r.namespaceId, ""))
	}
	// to determine the Content-Type header
	localVarHTTPContentTypes := []string{}

	// set Content-Type header
	localVarHTTPContentType := selectHeaderContentType(localVarHTTPContentTypes)
	if localVarHTTPContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHTTPContentType
	}

	// to determine the Accept header
	localVarHTTPHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHTTPHeaderAccept := selectHeaderAccept(localVarHTTPHeaderAccepts)
	if localVarHTTPHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHTTPHeaderAccept
	}
	if r.ctx != nil {
		// API Key Authentication
		if auth, ok := r.ctx.Value(ContextAPIKeys).(map[string]APIKey); ok {
			if apiKey, ok := auth["ApiKeyAuth"]; ok {
				var key string
				if apiKey.Prefix != "" {
					key = apiKey.Prefix + " " + apiKey.Key
				} else {
					key = apiKey.Key
				}
				localVarHeaderParams["Authorization"] = key
			}
		}
	}
	req, err := a.client.prepareRequest(r.ctx, localVarPath, localVarHTTPMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, formFiles)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHTTPResponse, err := a.client.callAPI(req)
	if err != nil || localVarHTTPResponse == nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHTTPResponse.Body)
	localVarHTTPResponse.Body.Close()
	localVarHTTPResponse.Body = ioutil.NopCloser(bytes.NewBuffer(localVarBody))
	if err != nil {
		return localVarReturnValue, localVarHTTPResponse, err
	}

	if localVarHTTPResponse.StatusCode >= 300 {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: localVarHTTPResponse.Status,
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	err = a.client.decode(&localVarReturnValue, localVarBody, localVarHTTPResponse.Header.Get("Content-Type"))
	if err != nil {
		newErr := &GenericOpenAPIError{
			body:  localVarBody,
			error: err.Error(),
		}
		return localVarReturnValue, localVarHTTPResponse, newErr
	}

	return localVarReturnValue, localVarHTTPResponse, nil
}
