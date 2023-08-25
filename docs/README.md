# PDS Test Suites

This repository is the collection of different test suites for PDS. Each test suite should create its own resources in
Control Plane and Target Cluster and thus avoid conflict of resources when running asynchronously.

## Building Test Suite

Each test suite can be built as go test binary which than can be invoked to run tests. Run `make all` to build all the
available test suites

## Running Test Suite

### Locally

A test suite can be executed locally by invoking the test binary on the local machine.

```shell
# To understand the test parameters
./bin/${SUITE}.test --help

# To run the tests
./bin/${SUITE}.tests --flags
```

### Inside Target Cluster

Test suites can be executed as containers in any kubernetes cluster. We have placed the config files in `config/` directory
that can be applied in the target cluster and test suite(s) will start running in parallel as kubernetes jobs.

### Pre-requisite

Bring a target cluster already registered to PDS.

#### Prepare the environment file

Make use of `./config/environment.template` and update the test parameters

#### Generate Kustomization config

```shell
# Understand the config parameters
./config/helper.sh help

# Generate the Kustomization Config
./config/helper.sh config

Eg -
TESTS=iam ENVIRONMENT_FILE=<env.file> ./config/helper.sh config >> ./config/kustomization.yml

Using a dev test image -
TESTS=iam ENVIRONMENT_FILE=<env.file> TEST_IMAGE_NAME="myimage" TEST_IMAGE_VERSION="version" \
./config/helper.sh config >> ./config/kustomization.yml
```

#### Apply the kustomization resources

```shell
kubectl apply -k ./config
```

All the test jobs will come up inside namespace `pds-integration-tests`. Test logs can be examined by looking into pods
logs

```shell
kubectl get jobs -n pds-integration-tests
kubectl get pods -n pds-integration-tests
```
