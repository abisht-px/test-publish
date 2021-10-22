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

## Manually triggering test on GitHub

A test run can be manually requested via the
[GitHub UI](https://github.com/portworx/pds-integration-test/actions/workflows/ci-test.yml)
or [GitHub CLI](https://github.com/cli/cli) using the command:

```shell
gh workflow run ci-test.yml
```

This runs the workflow against `master` using the default parameters.

To override the parameters, e.g. to run against the currently checked out branch
(provided it's also present on remote), using a `test` environment on `master` of
[pds-infra](https://github.com/portworx/pds-infra), run:

```shell
gh workflow run --ref $(git rev-parse --abbrev-ref HEAD) ci-test.yml -f branch=master -f environment=test
```

See [CI Test](./.github/workflows/ci-test.yml) inputs for possible overrides.
