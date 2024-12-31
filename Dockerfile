ARG VERSION

FROM golang:1.22.5 AS common-build
ARG TARGETOS TARGETARCH
WORKDIR /source
ENV CGO_ENABLED=1
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

FROM --platform=linux/amd64 common-build AS linux-amd64-builder
RUN dpkg --add-architecture amd64 \
    && apt-get update \
    && apt-get install -y --no-install-recommends gcc-x86-64-linux-gnu libc6-dev-amd64-cross
ENV CC=x86_64-linux-gnu-gcc

FROM --platform=linux/amd64 linux-amd64-builder AS build-linux-amd64
ARG VERSION
COPY . .
RUN go build -ldflags="-extldflags=-static -X main.version=$VERSION" -o exporter ./cmd/exporter

FROM --platform=linux/arm64 common-build AS build-linux-arm64
ARG VERSION
COPY . .
RUN go build -ldflags="-extldflags=-static -X main.version=$VERSION" -o exporter ./cmd/exporter

FROM alpine:3.21.0 AS final
WORKDIR /
VOLUME [ "/data" ]
ENTRYPOINT ["/exporter"]

FROM --platform=linux/arm64 final AS arm64
COPY --from=build-linux-arm64 /source/exporter .

FROM --platform=linux/amd64 final AS amd64
COPY --from=build-linux-amd64 /source/exporter .
