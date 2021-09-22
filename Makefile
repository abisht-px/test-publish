test:
	go test ./... -v

vendor:
	go mod tidy
	go mod vendor

.PHONY: test vendor
