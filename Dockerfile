ARG GOLANG_VERSION=1.17
ARG VERSION=0.1.0

FROM golang:${GOLANG_VERSION} AS builder
ARG VERSION
# enable Go modules support
ENV GO111MODULE=on
ENV CGO_ENABLED=0

WORKDIR sbercdn-exporter

# Copy src code from the host and compile it
COPY apiclient ./apiclient
COPY common ./common
COPY collector ./collector
COPY go.* *.go ./
RUN set -xe && \
    go mod tidy && \
    go build -a -trimpath -ldflags "-X main.Version=$VERSION -w" -o /sbercdn-exporter

###
FROM scratch
LABEL maintainer="o.marin@rabota.ru"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /sbercdn-exporter /bin/
EXPOSE 9921/tcp
ENTRYPOINT ["/bin/sbercdn-exporter"]
