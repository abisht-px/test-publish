package portworx

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/portworx/pds-integration-test/internal/controlplane"

	"github.com/hashicorp/go-multierror"
)

type PXCloudCredential struct {
	ID            string `json:"credential_id"`
	Name          string `json:"name"`
	Bucket        string `json:"bucket"`
	AwsCredential struct {
		AccessKey string `json:"access_key"`
		Endpoint  string `json:"endpoint"`
		Region    string `json:"region"`
	} `json:"aws_credential"`
}

type AWSCredentialsRequest struct {
	AccessKey string `json:"access_key,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
	Region    string `json:"region,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}

type CreateCredentialsRequest struct {
	AwsCredential *AWSCredentialsRequest `json:"aws_credential,omitempty"`
	Bucket        string                 `json:"bucket,omitempty"`
	Name          string                 `json:"name"`
}

// GetPXCloudCredential gets single Portworx cloud credential.
func (p *Portworx) GetPXCloudCredential(ctx context.Context, credentialID string) (PXCloudCredential, error) {
	cloudCredentialJSON, err := p.buildPXAPIRequest(p.restClient.Get(), "v1/credentials/inspect/"+credentialID).Do(ctx).Raw()
	if err != nil {
		return PXCloudCredential{}, err
	}

	var cloudCredential PXCloudCredential
	err = json.Unmarshal(cloudCredentialJSON, &cloudCredential)
	if err != nil {
		return PXCloudCredential{}, err
	}
	return cloudCredential, nil
}

// ListPXCloudCredentials lists all Portworx cloud credentials.
func (p *Portworx) ListPXCloudCredentials(ctx context.Context) ([]PXCloudCredential, error) {
	// listPXCredentialsResponse is response from the Portworx API containing only a list of credential IDs.
	type listPXCredentialsResponse struct {
		CredentialIDs []string `json:"credential_ids"`
	}
	credentialsJSON, err := p.buildPXAPIRequest(p.restClient.Get(), "v1/credentials").Do(ctx).Raw()
	if err != nil {
		return nil, err
	}

	var credentialsResponse listPXCredentialsResponse
	err = json.Unmarshal(credentialsJSON, &credentialsResponse)
	if err != nil {
		return nil, err
	}

	// For each cloud credential ID, let's fetch the full object.
	var cloudCredentials []PXCloudCredential
	for _, credentialID := range credentialsResponse.CredentialIDs {
		cloudCredential, err := p.GetPXCloudCredential(ctx, credentialID)
		if err != nil {
			return nil, err
		}
		cloudCredentials = append(cloudCredentials, cloudCredential)
	}
	return cloudCredentials, nil
}

// DeletePXCloudCredential deletes single Portworx cloud credential by the ID.
func (p *Portworx) DeletePXCloudCredential(ctx context.Context, cloudCredentialID string) error {
	_, err := p.buildPXAPIRequest(p.restClient.Delete(), "v1/credentials/"+cloudCredentialID).Do(ctx).Raw()
	return err
}

// DeletePXCloudCredentials deletes all Portworx credentials. Used in the test cleanup.
func (p *Portworx) DeletePXCloudCredentials(ctx context.Context) error {
	credentials, err := p.ListPXCloudCredentials(ctx)
	if err != nil {
		return err
	}

	for _, credential := range credentials {
		if strings.HasPrefix(credential.Name, "pdscreds-") {
			deleteErr := p.DeletePXCloudCredential(ctx, credential.ID)
			if deleteErr != nil {
				err = multierror.Append(err, deleteErr)
			}
		}
	}
	return nil
}

func (p *Portworx) FindCloudCredentialByName(ctx context.Context, name string) (*PXCloudCredential, error) {
	credentials, err := p.ListPXCloudCredentials(ctx)
	if err != nil {
		return nil, err
	}
	for _, credential := range credentials {
		if credential.Name == name {
			return &credential, nil
		}
	}
	return nil, fmt.Errorf("cloud credential '%s' not found", name)
}

func (p *Portworx) CreatePXCloudCredentialsForS3(ctx context.Context, name, bucket string, s3 controlplane.S3Credentials) error {
	requestBody := CreateCredentialsRequest{
		Name:   name,
		Bucket: bucket,
		AwsCredential: &AWSCredentialsRequest{
			AccessKey: s3.AccessKey,
			SecretKey: s3.SecretKey,
			Region:    "us-west-2",
			Endpoint:  s3.Endpoint,
		},
	}

	rawBody, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	_, err = p.buildPXAPIRequest(p.restClient.Post(), "v1/credentials").Body(rawBody).Do(ctx).Raw()
	if err != nil {
		return err
	}
	return nil
}
