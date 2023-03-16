FROM golang:1.19.7-alpine as builder

ARG VERSION=dev

ENV GO111MODULE=on
ENV CGO_ENABLED=0

ENV SRC_PATH=github.com/mmohamed/kubernetes-volume-usage-prometheus-metric

RUN apk add -U --no-cache git ca-certificates && \
    mkdir -p /go/src/${SRC_PATH}

WORKDIR /go/src/${SRC_PATH}

COPY go.mod go.mod
COPY go.sum go.sum
COPY cmd cmd
COPY pkg pkg

RUN set -eux \
    && cd cmd/volume-usage-exporter \
    && go build -ldflags="-X 'main.version=${VERSION}'" \
    -o /go/src/${SRC_PATH}/volume-usage-exporter

FROM scratch

ENV SRC_PATH=github.com/mmohamed/kubernetes-volume-usage-prometheus-metric

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/${SRC_PATH}/volume-usage-exporter /volume-usage-exporter

ENTRYPOINT ["/volume-usage-exporter"]