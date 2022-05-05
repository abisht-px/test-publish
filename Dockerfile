# Build the test binary.
FROM golang:1.16.14 as builder

WORKDIR /workspace

# Dependencies.
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/
COPY Makefile Makefile

# Source.
COPY test/ test/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -c -o pds-test ./test

# Use distroless as minimal base image to package the test binary.
# Refer to https://github.com/GoogleContainerTools/distroless for more details.
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/pds-test .

USER nonroot:nonroot

ENTRYPOINT ["/pds-test"]
