# PRERELEASE defines semver prerelease tag based on the state of the current git tree.
# For a clean tree this is simply short commit hash of HEAD, e.g., "4be01eb".
# Dirty tree has "-dirty" suffix added, e.g., "4be01eb-dirty".
GOIMPORTS_BIN=bin/goimports.out
GOIMPORTS_CMD=./$(GOIMPORTS_BIN) -local github.com/portworx/pds-integration-test -w
PRERELEASE = $(shell git describe --match= --always --dirty)

# TAG for the test docker image, e.g., "dev-4be01eb-dirty".
IMG_TAG = dev-$(PRERELEASE)
IMG_REPO = docker.io

# Image URL to use all building/pushing image targets
IMG = $(IMG_REPO)/portworx/pds-integration-test:$(IMG_TAG)

.PHONY: test vendor lint docker-build docker-push fmt

fmt:
	go build -o $(GOIMPORTS_BIN) golang.org/x/tools/cmd/goimports
	find . -name "*.go" -type f -exec $(GOIMPORTS_CMD) {} \; -o -path './vendor' -prune

test:
	go test ./... -v

vendor:
	go mod tidy
	go mod vendor

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run

mdlint:
	docker run --rm -v $$PWD:/workdir davidanson/markdownlint-cli2 "**/*.md" "#vendor"

docker-build:
	docker build . -t ${IMG}

docker-push:
	docker push ${IMG}
