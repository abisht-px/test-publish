# PDS Integration Tests

End-to-end test definitions for PDS.

## Control Plane cluster

The control plane is created from scratch on every test run using [kind](https://kind.sigs.k8s.io/).
PDS components are then deployed based on the [pds-infra](https://github.com/portworx/pds-infra)
definitions on a specified branch (`ci` by default).

## Target cluster

For now, we are using a long-lived target cluster, which gets merely injected into the tests. The connection parameters are stored in
[vault](https://secret-service.inf-cloud-support.purestorage.com/ui/vault/secrets/secret/show/engineering/portworx/pds/integration-test).

This means we have to be more attentive with cleaning up any stray resources after the tests and have to manually monitor and maintain the health of the cluster.
