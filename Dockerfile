# Build the proxy binary.
FROM golang:1.16.6 as builder

WORKDIR /workspace

# Dependencies.
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/
COPY Makefile Makefile

# Source.
COPY test/ test/
COPY setup/ setup/

CMD go test ./test -v

