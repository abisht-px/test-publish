FROM golang:1.20 as builder

WORKDIR /workspace

# Dependencies.
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/

COPY cmd/ ./cmd

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/go-testify-report ./cmd/tools/report
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/go-test-doc ./cmd/tools/doc

FROM gcr.io/distroless/static-debian11

COPY --from=builder /workspace/bin/* /bin/

WORKDIR /reports

CMD [""]