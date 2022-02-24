test:
	go test ./... -v

vendor:
	go mod tidy
	go mod vendor

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run 

.PHONY: test vendor lint
