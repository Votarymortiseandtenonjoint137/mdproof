# mdproof sandbox — Go development and test environment.
# No frontend dependencies needed (pure Go CLI).
ARG GO_VERSION=1.25.5

FROM golang:${GO_VERSION}-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash git ca-certificates curl jq sudo \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

COPY go.mod ./
RUN go mod download
RUN mkdir -p /go/build-cache /go/pkg/mod && chmod -R 0777 /go/build-cache /go/pkg/mod

CMD ["bash", "-c", "./scripts/test.sh"]
