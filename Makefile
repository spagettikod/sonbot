VERSION=1.0.0
OUTPUT=build
.PHONY: build_linux build_macos pkg_linux pkg_macos all default clean setup docker test

default: help

##@ Commands
help:
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make <COMMAND>\033[36m\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


clean:						## Clean up folders and files created by this Makefile
	@rm -rf $(OUTPUT)

test:						## Run all tests
	@go test .

setup:
	@mkdir -p $(OUTPUT)/linux
	@mkdir -p $(OUTPUT)/macos

linux:						## Build container for linux/amd64
	@docker build --build-arg VERSION=$(VERSION) --platform linux/amd64 --file Dockerfile-exporter -t registry.spagettikod.se:8443/tpir/exporter:latest --target amd64 --push .

linux_bin: clean setup		## Build linux/amd64 binary
	@docker build --build-arg VERSION=$(VERSION) --platform linux/amd64 --file Dockerfile-exporter -t registry.spagettikod.se:8443/tpir/exporter:latest --target amd64 --output=$(OUTPUT)/linux .

macos:						## Build MacOS container for
	@docker build --build-arg VERSION=$(VERSION) --platform linux/arm64 --file Dockerfile-exporter -t registry.spagettikod.se:8443/tpir/exporter:latest --target arm64 --push --load .

macos_bin: clean setup		## Build MacOS binary
	@env CGO_ENABLED=1 go build -ldflags="-X cmd.exporter.main.version=$(VERSION)" -o $(OUTPUT)/macos/exporter ./cmd/exporter

pkg_linux: build_linux
	@tar -C $(OUTPUT)/linux -czf $(OUTPUT)/tpir$(VERSION).linux-amd64.tar.gz tpir

pkg_macos: build_macos
	@tar -C $(OUTPUT)/macos -czf $(OUTPUT)/tpir$(VERSION).macos-arm64.tar.gz tpir

all: clean pkg_linux pkg_macos
