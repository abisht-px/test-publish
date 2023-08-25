# Build the test binary.
FROM golang:1.19.4 as builder

WORKDIR /workspace

# Dependencies.
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/
COPY Makefile Makefile

# Source.
COPY test/ test/
COPY internal/ internal/

RUN wget -qO- https://github.com/jstemmer/go-junit-report/releases/download/v2.0.0/go-junit-report-v2.0.0-linux-amd64.tar.gz | tar xzv

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go test -c -o pds-test ./test

# Use alpine as minimal base image to package the test binary.
FROM alpine:3.17.3

ENV PDS_JUNIT_REPORT_FILEPATH=report.xml
ENV PDS_TEST_ARGS=""

WORKDIR /
COPY --from=builder /workspace/pds-test .
COPY --from=builder /workspace/go-junit-report .

ENTRYPOINT ["/bin/sh", "-c", "/pds-test -test.v ${PDS_TEST_ARGS} | /go-junit-report -iocopy -set-exit-code -out $PDS_JUNIT_REPORT_FILEPATH"]