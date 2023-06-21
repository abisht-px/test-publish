# PDS Integration Tests

End-to-end test definitions for PDS.

The test is intended to be run inside a [Jenkins pipeline](https://jenkins.pwx.dev.purestorage.com/job/PDS/job/integration-test/)
against a Docker image.
Control plane and a target cluster is required, but not spawned by the test - these must be provisioned separately
and their connection details injected via environment variables.

For instructions how to run tests locally, see [Running tests locally](#running-tests-locally).

## Environment Configuration

| Key                        | Default value                              | Description                                                                                                       |
|----------------------------|--------------------------------------------|-------------------------------------------------------------------------------------------------------------------|
| CONTROL_PLANE_ADDRESS          |                                            | Domain name or ip address of the publicly accessible control plane endpoint.                                                             |
| TARGET_CLUSTER_KUBECONFIG  |                                            | Path to the kubeconfig file that allows the test to access the target cluster.                                    |
| PDS_ACCOUNT_NAME           | PDS Integration Test                       | Name of the PDS account to be used for the test.                                                                  |
| PDS_TENANT_NAME            | Default                                    | Name of the PDS tenant to be used for the test.                                                                   |
| PDS_DEPTARGET_NAME         | PDS Integration Test Cluster \<timestamp\> | Name of the PDS deployment target to be used for the test. Can be left empty if chart installation isn't skipped. |
| PDS_SERVICE_ACCOUNT_NAME   | Default-AgentWriter                        | Name of the PDS service account to be used for agent installation.                                                |
| PDS_PROJECT_NAME           | Default                                    | Name of the PDS project to be used for the test.                                                                  |
| PDS_NAMESPACE_NAME         | dev                                        | Name of the PDS namespace to be used for the test.                                                                |
| PDS_BACKUPTARGET_BUCKET    |                                            | Bucket name for the S3 or S3 compatible service.                                                                  |
| PDS_BACKUPTARGET_REGION    |                                            | Region of the bucket, required for S3.                                                                            |
| PDS_S3CREDENTIALS_ENDPOINT | s3.amazonaws.com                           | Base path for the AWS S3 endpoint.                                                                                |
| PDS_HELM_CHART_VERSION     | configured by control plane                | PDS Helm chart version. Use "0" to skip PDS chart installation.                                                   |

### Secrets

| Key                           | Description                                                                         |
|-------------------------------|-------------------------------------------------------------------------------------|
| SECRET_TOKEN_ISSUER_URL       | Base URL of the token issuer that can provide us with a bearer token.               |
| SECRET_ISSUER_CLIENT_ID       | ClientID to be used when querying the token issuer.                                 |
| SECRET_ISSUER_CLIENT_SECRET   | Secret to authenticate with the issuer.                                             |
| SECRET_PDS_USERNAME           | Username of a PDS user on the control plane. This user must already be pre-created. |
| SECRET_PDS_PASSWORD           | Password corresponding to the PDS user.                                             |
| SECRET_PDS_TOKEN              | User api token (can have custom expiration date)                                    |
| PDS_S3CREDENTIALS_ACCESSKEY   | AWS access key used for the pds backup credentials.                                 |
| PDS_S3CREDENTIALS_SECRETKEY   | AWS secret key used for the pds backup credentials.                                 |
| SECRET_PDS_AUTH_USER_USERNAME | Username of a existing PDS user on the control plane for the Auth test.             |
| SECRET_PDS_AUTH_USER_PASSWORD | Password corresponding to the PDS Auth test user.                                   |
| SECRET_PDS_AUTH_USER_TOKEN    | Auth test user api token (can have custom expiration date)                          |

NOTE:

* `SECRET_PDS_TOKEN` can be used for auth instead of user/password one.
* `SECRET_PDS_AUTH_USER_TOKEN` can be used for auth instead of user/password one.
* The Auth User must be existing user in the control plane environment. During the test the role bindings for the
user are changed in the selected Account. And at the end of the test there will be no role for the user. You should use
different user than the main for the tests! Also the test cannot be run in parallel, as there will be race condition on
the bindings.

Make sure you have added helm chart manually to your local.
Since minihelm looks for this entry it fails if you do not define this.
See this slack [thread](https://purestorage.slack.com/archives/C04CQSSMFPC/p1669717983272019)

Add the following entry to `/home/nonroot/.config/helm/repositories.yaml` (for Mac users `~/Library/Preferences/helm/repositories.yaml`):

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

## Running on Jenkins

The [integration test job] on Jenkins is configured to be compatible with the test images.

You need to be on VPN to access Jenkins.

Use [`Build with parameters`](https://jenkins.pwx.dev.purestorage.com/job/PDS/job/integration-test/build) to trigger a run.

## Running tests locally

In order to run test locally, you can optionally create a `.env` file at the root of the project to override your
environment variables in one place. A template of the `.env` file:

```dotenv
# Control plane and target cluster config.
CONTROL_PLANE_ADDRESS='<domain name or ip address>'
TARGET_CLUSTER_KUBECONFIG=''

# OIDC config.
SECRET_TOKEN_ISSUER_URL='<environment/px-central/token-issuer secret in AWS secret manager>'
SECRET_ISSUER_CLIENT_ID='<environment/px-central/token-issuer secret in AWS secret manager>'
SECRET_ISSUER_CLIENT_SECRET='<environment/px-central/token-issuer secret in AWS secret manager>'

# Authentication credentials for PDS.
SECRET_PDS_USERNAME='<environment/px-central/admin-credentials secret in AWS secret manager>'
SECRET_PDS_PASSWORD='<environment/px-central/admin-credentials secret in AWS secret manager>'

# Disable Helm chart installation and specify deployment target name.
PDS_HELM_CHART_VERSION=0
PDS_DEPTARGET_NAME=''

# (optional) S3 Backup target configuration.
PDS_BACKUPTARGET_BUCKET=''
PDS_BACKUPTARGET_REGION=''
PDS_S3CREDENTIALS_ACCESSKEY=''
PDS_S3CREDENTIALS_SECRETKEY=''
```

The following command only runs a single test from the suite.

```bash
go test -test.v ./... -testify.m {put_test_name_here}

#for instance
go test -parallel 1 -test.v ./... -testify.m TestDataService_UpdateImage

# run 2 parallel tests at once:
go test -parallel 2 -test.v ./... -testify.m TestDataService_UpdateImage

# as parallel uses GOMAXPROCS value, the following is the same:
export GOMAXPROCS=2
go test -test.v ./... -testify.m TestDataService_UpdateImage
```

## Development

### Parallel tests

Testify isn’t really safe for parallel test execution within the same suite ([link](https://github.com/stretchr/testify/pull/1322)).

All the tests share the same testing.T instance, so tests start affecting each other in strange
 ways, and the cleanups can start racing with the suite cleanup, etc.

But we still could mark all subtests as parallel within the parent test, this requires:

* use `s.T().Run(description, func (t *testing.T) {` instead of
`s.Run(description, func() {` in order to explicitly get the new testing.T instance;
* any helpers and assertions can’t use the suite’s top-level `s.T()` instance,
but they have to accept a `*testing.T` parameter. Then the callers either pass
the separate subtest `t`, or `s.T()` if they’re not inside a subtest.
