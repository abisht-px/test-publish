# PDS Integration Tests

End-to-end test definitions for PDS.

The test is intended to be run inside a Jenkins pipeline against a Docker image.
Control plane and a target cluster is required, but not spawned by the test - these must be provisioned separately
and their connection details injected via environment variables.

Running locally (`make test`) is also possible if:

- your environment is authorized to talk to the control plane and target cluster.
- you provide the required environment variables to the test.

## Environment Configuration

| Key                       | Description                                                                           |
|---------------------------|---------------------------------------------------------------------------------------|
| CONTROL_PLANE_API         | URL of the publicly accessible control plane PDS API.                                 |
| TARGET_CLUSTER_KUBECONFIG | Path to the kubeconfig file that allows the test to access the target cluster.        |
