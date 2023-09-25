# PRERELEASE defines semver prerelease tag based on the state of the current git tree.
# For a clean tree this is simply short commit hash of HEAD, e.g., "4be01eb".
# Dirty tree has "-dirty" suffix added, e.g., "4be01eb-dirty".
GOIMPORTS_BIN=bin/goimports.out
GOIMPORTS_CMD=./$(GOIMPORTS_BIN) -local github.com/portworx/pds-integration-test -w
PRERELEASE = $(shell git describe --match= --always --dirty)

# TAG for the test docker image, e.g., "dev-4be01eb-dirty".
IMG_TAG = dev-$(PRERELEASE)
IMG_REPO = docker.io/portworx

# Image URL to use all building/pushing image targets
IMG = $(IMG_REPO)/pds-integration-test:$(IMG_TAG)
SUITES_IMG = $(IMG_REPO)/pds-integration-test-suites:$(IMG_TAG)
CONFIG_IMG = $(IMG_REPO)/pds-integration-test-config:$(IMG_TAG)
TOOLS_IMG = $(IMG_REPO)/pds-integration-test-tools:$(IMG_TAG)

DOC_PKGS = "backup,backupjob,capabilities,copilot,dataservices,deployment,iam,namespace,portworxcsi,reporting,restore,targetcluster,tls"
DOC_FORMAT = "json"

# Default testrail values for section and project id
sectionID = 9074
projectID = 1

.PHONY: test vendor lint docker-build docker-push fmt doc

all: build fmt lint

build-testreport:
	go build -o ./bin/testreport ./cmd/report

build:
	go test -c -o ./bin/register.test ./suites/register
	go test -c -o ./bin/iam.test ./suites/iam
	go test -c -o ./bin/namespace.test ./suites/namespace
	go test -c -o ./bin/backup.test ./suites/backup
	go test -c -o ./bin/backupjob.test ./suites/backupjob
	go test -c -o ./bin/restore.test ./suites/restore
	go test -c -o ./bin/deployment.test ./suites/deployment
	go test -c -o ./bin/portworxcsi.test ./suites/portworxcsi
	go test -c -o ./bin/targetcluster.test ./suites/targetcluster
	go test -c -o ./bin/reporting.test ./suites/reporting
	go test -c -o ./bin/capabilities.test ./suites/capabilities
	go test -c -o ./bin/dataservices.test ./suites/dataservices
	go test -c -o ./bin/tls.test ./suites/tls
	go test -c -o ./bin/copilot.test ./suites/copilot

build-%:
	go test -c -o ./bin/$(*).test ./suites/$(*)

fmt:
	go build -o $(GOIMPORTS_BIN) golang.org/x/tools/cmd/goimports
	find . -name "*.go" -type f -exec $(GOIMPORTS_CMD) {} \; -o -path './vendor' -prune

test:
	go test ./... -v

doc:
	@go run ./cmd/tools/doc --publish=$(publish) --baseDir="./suites" --pkgs=$(DOC_PKGS) --format=$(DOC_FORMAT) --testrailUserName=$(testrailUserName) --testrailAPIKey=$(testrailAPIKey) --sectionID=$(sectionID) --projectID=$(projectID)

doc-old:
	@go run ./cmd/tools/doc --baseDir="." --pkgs="test" --format=$(DOC_FORMAT)

vendor:
	go mod tidy
	go mod vendor

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix

mdlint:
	docker run --rm -v $$PWD:/workdir davidanson/markdownlint-cli2 "**/*.md" "#vendor"

docker: docker-build-all docker-push-all
docker-build-all: docker-build docker-build-suites docker-build-config docker-build-tools
docker-push-all: docker-push docker-push-suites docker-push-config docker-push-tools

docker-build:
	docker build . -t ${IMG}

docker-build-suites:
	docker build . -f Dockerfile.suites -t ${SUITES_IMG}

docker-build-config:
	docker build ./config/. -t ${CONFIG_IMG}

docker-build-tools:
	docker build -f ./cmd/tools/Dockerfile -t ${TOOLS_IMG} .

docker-push:
	docker push ${IMG}

docker-push-suites:
	docker push ${SUITES_IMG}

docker-push-config:
	docker push ${CONFIG_IMG}

docker-push-tools:
	docker push ${TOOLS_IMG}

run-register:
	./bin/register.test -controlPlaneAPI=${CONTROL_PLANE_API} \
	-accountName="${ACCOUNT_NAME}" \
	-tenantName=${TENANT_NAME} \
	-projectName=${PROJECT_NAME} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsHelmChartVersion="1.19.0" \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-deploymentTargetName=${DEPLOYMENT_TARGET_NAME} \
	-registerOnly=true \
	-dataServicesTLSEnabled=true \
	-test.failfast \
	-test.v

run-namespaces:
	./bin/namespace.test -controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsHelmChartVersion="" \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-test.failfast \
	-test.v

run-backup:
	./bin/backup.test -controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsHelmChartVersion="0" \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-awsAccessKey=${AWS_ACCESS_KEY} \
	-awsSecretKey=${AWS_SECRET_KEY} \
	-awsS3BucketName=${AWS_S3_BUCKET_NAME} \
	-deploymentTargetName=${DEPLOYMENT_TARGET_NAME} \
	-test.failfast \
	-test.v

run-backupjob:
	./bin/backupjob.test -controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsHelmChartVersion="0" \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-awsAccessKey=${AWS_ACCESS_KEY} \
	-awsSecretKey=${AWS_SECRET_KEY} \
	-awsS3BucketName=${AWS_S3_BUCKET_NAME} \
	-deploymentTargetName=${DEPLOYMENT_TARGET_NAME} \
	-test.failfast \
	-test.v

run-iam:
	./bin/iam.test -controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-accountName="${ACCOUNT_NAME}" \
	-pdsToken=${PDS_API_TOKEN} \
	-authUserName=${PDS_AUTH_USER_NAME} \
	-authPassword=${PDS_AUTH_USER_PASSWORD} \
	-additionalAccounts="${ADDITIONAL_PDS_ACCOUNTS}" \
	-test.failfast \
	-test.v

run-deployment:
	./bin/deployment.test -controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsHelmChartVersion="0" \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-accountName="PDS Functional tests" \
	-deploymentTargetName=${DEPLOYMENT_TARGET_NAME} \
	-test.run="TestDeploymentTestSuite/TestDeploymentStatuses_Available" \
	-test.failfast \
	-test.v

run-portworxcsi:
	./bin/portworxcsi.test --controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-pdsHelmChartVersion="0" \
	-accountName="${ACCOUNT_NAME}" \
	-tenantName=${TENANT_NAME} \
	-projectName=${PROJECT_NAME} \
	-deploymentTargetName=${DEPLOYMENT_TARGET_NAME} \
	-test.failfast \
	-test.v

run-targetcluster:
	./bin/targetcluster.test --controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-pdsHelmChartVersion="0" \
	-accountName="${ACCOUNT_NAME}" \
	-tenantName=${TENANT_NAME} \
	-projectName=${PROJECT_NAME} \
	-deploymentTargetName=${DEPLOYMENT_TARGET_NAME} \
	-test.failfast \
	-test.v


run-reporting:
	./bin/reporting.test --controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-pdsHelmChartVersion="0" \
	-accountName="${ACCOUNT_NAME}" \
	-tenantName=${TENANT_NAME} \
	-projectName=${PROJECT_NAME} \
	-deploymentTargetName=${DEPLOYMENT_TARGET_NAME} \
	-test.failfast \
	-test.v

run-dataservices:
	./bin/dataservices.test -controlPlaneAPI="${CONTROL_PLANE_API}" \
	-issuerClientSecret="${ISSUER_CLIENT_SECRET}" \
	-accountName="${ACCOUNT_NAME}" \
	-tenantName=${TENANT_NAME} \
	-projectName=${PROJECT_NAME} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsHelmChartVersion="0" \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-awsAccessKey=${AWS_ACCESS_KEY} \
  	-awsSecretKey=${AWS_SECRET_KEY} \
  	-awsS3BucketName=${AWS_S3_BUCKET_NAME} \
  	-deploymentTargetName=${DEPLOYMENT_TARGET_NAME} \
  	-test.failfast \
  	-test.run="TestDataservicesSuite" \
  	-test.v

run-capabilities:
	./bin/capabilities.test --controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-pdsHelmChartVersion="1.20.1" \
	-accountName="${ACCOUNT_NAME}" \
	-tenantName=${TENANT_NAME} \
	-projectName=${PROJECT_NAME} \
	-deploymentTargetName=${DEPLOYMENT_TARGET_NAME} \
	-test.failfast \
	-test.v

run-tls:
	./bin/tls.test -controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsHelmChartVersion="0" \
	-pdsToken=${PDS_API_TOKEN} \
	-targetClusterKubeconfig=${TC_KUBECONFIG} \
	-accountName=${ACCOUNT_NAME} \
	-deploymentTargetName=${DEPLOYMENT_TARGET_NAME} \
	-dataServicesTLSEnabled=true \
	-test.run="TestTLSSuite" \
	-test.failfast \
	-test.v

run-copilot:
	./bin/copilot.test -controlPlaneAPI=${CONTROL_PLANE_API} \
	-issuerClientSecret=${ISSUER_CLIENT_SECRET} \
	-issuerClientID=${ISSUER_CLIENT_ID} \
	-issuerTokenURL=${ISSUER_TOKEN_URL} \
	-pdsToken=${PDS_API_TOKEN} \
	-test.run="TestCopilotTestSuite" \
	-test.failfast \
	-test.v
