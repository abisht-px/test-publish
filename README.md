# PDS Integration Tests

End-to-end test definitions for PDS.

The test is intended to be run inside a Jenkins pipeline against a Docker image.
Control plane and a target cluster is required, but not spawned by the test - these must be provisioned separately
and their connection details injected via environment variables.

Running locally (`make test`) is also possible if:

- your environment is authorized to talk to the control plane and target cluster.
- you provide the required environment variables to the test.

## Environment Configuration

| Key                        | Default value                | Description                                                                    |
|----------------------------|------------------------------|--------------------------------------------------------------------------------|
| CONTROL_PLANE_API          |                              | URL of the publicly accessible control plane PDS API.                          |
| TARGET_CLUSTER_KUBECONFIG  |                              | Path to the kubeconfig file that allows the test to access the target cluster. |
| PDS_ACCOUNT_NAME           | PDS Integration Test         | Name of the PDS account to be used for the test.                               |
| PDS_TENANT_NAME            | Default                      | Name of the PDS tenant to be used for the test.                                |
| PDS_DEPTARGET_NAME         | PDS Integration Test Cluster | Name of the PDS deployment target to be used for the test.                     |
| PDS_SERVICE_ACCOUNT_NAME   | Default-AgentWriter          | Name of the PDS service account to be used for agent installation.             |
| PDS_PROJECT_NAME           | Default                      | Name of the PDS project to be used for the test.                               |
| PDS_NAMESPACE_NAME         | dev                          | Name of the PDS namespace to be used for the test.                             |
| PX_NAMESPACE_NAME          | kube-system                  | Name of the PX namespace to be used for the test.                              |
| PDS_BACKUPTARGET_BUCKET    |                              | Bucket name for the S3 or S3 compatible service.                               |
| PDS_BACKUPTARGET_REGION    |                              | Region of the bucket, required for S3.                                         |
| PDS_S3CREDENTIALS_ENDPOINT | s3.amazonaws.com             | Base path for the AWS S3 endpoint.                                             |

### Secrets

| Key                         | Description                                                                         |
|-----------------------------|-------------------------------------------------------------------------------------|
| SECRET_TOKEN_ISSUER_URL     | Base URL of the token issuer that can provide us with a bearer token.               |
| SECRET_ISSUER_CLIENT_ID     | ClientID to be used when querying the token issuer.                                 |
| SECRET_ISSUER_CLIENT_SECRET | Secret to authenticate with the issuer.                                             |
| SECRET_PDS_USERNAME         | Username of a PDS user on the control plane. This user must already be pre-created. |
| SECRET_PDS_PASSWORD         | Password corresponding to the PDS user.                                             |
| SECRET_PDS_TOKEN            | User api token (can have custom expiration date)                                    |
| PDS_S3CREDENTIALS_ACCESSKEY | AWS access key used for the pds backup credentials.                                 |
| PDS_S3CREDENTIALS_SECRETKEY | AWS secret key used for the pds backup credentials.                                 |

NOTE: `SECRET_PDS_TOKEN` can be used for auth instead of user/password one.

Make sure you have added helm chart manually to your local.
Since minihelm looks for this entry it fails if you do not define this.
See this slack [thread](https://purestorage.slack.com/archives/C04CQSSMFPC/p1669717983272019)

Add the following entry to `/home/nonroot/.config/helm/repositories.yaml`:

```bash
cat <<EOT >> /home/nonroot/.config/helm/repositories.yaml
apiVersion: ""
generated: "0001-01-01T00:00:00Z"
repositories:
- caFile: ""
  certFile: ""
  insecure_skip_tls_verify: false
  keyFile: ""
  name: pds
  pass_credentials_all: true
  password: <PX_PASSWORD>
  url: https://portworx.github.io/pds-charts
  username: <PX_USER>
EOT
```

## Example run

this example loads one deployment specification and starts all test cases

```bash
HELM_NAMESPACE='pds-system' \
TARGET_CLUSTER_KUBECONFIG='pds-tc-large-01.yaml' \
CONTROL_PLANE_API='https://staging.pds.portworx.com/api' \
SECRET_TOKEN_ISSUER_URL='https://apicentral.portworx.com/api' \
SECRET_ISSUER_CLIENT_ID='...' \
SECRET_ISSUER_CLIENT_SECRET='...' \
SECRET_PDS_USERNAME='...' \
SECRET_PDS_PASSWORD='...' \
PDS_ACCOUNT_NAME='Portworx' \
PDS_BACKUPTARGET_BUCKET='...' \
PDS_BACKUPTARGET_REGION='...' \
PDS_S3CREDENTIALS_ACCESSKEY='...' \
PDS_S3CREDENTIALS_ENDPOINT='...' \
PDS_S3CREDENTIALS_SECRETKEY='...' \
go test -test.v -test.run=Test ./...
```

the following command only runs a single test from the suite.

```bash
go test -test.v ./... -testify.m {put_test_name_here}

#for instance
go test -test.v ./... -testify.m TestDataService_UpdateImage
```
