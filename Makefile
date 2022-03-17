TEST_IMG=portworx/pds-integration-test:$(DOCKER_HUB_TAG)
test:
	go test ./... -v

vendor:
	go mod tidy
	go mod vendor

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run 

.PHONY: test vendor lint

container:
	@echo "Building container: docker build --tag $(TEST_IMG) -f Dockerfile ."
	sudo docker build --tag $(TEST_IMG) -f Dockerfile .
	sudo docker push $(TEST_IMG)