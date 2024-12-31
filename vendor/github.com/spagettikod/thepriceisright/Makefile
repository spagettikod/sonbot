VERSION=1.1.0
OUTPUT=build
.PHONY: build_linux build_macos pkg_linux pkg_macos all default clean setup docker test

default: test

clean:
	@rm -rf $(OUTPUT)

test:
	@go test .

setup:
	@mkdir -p $(OUTPUT)/linux
	@mkdir -p $(OUTPUT)/macos

build_linux: setup
	@env GOOS=linux GOARCH=amd64 go build -o $(OUTPUT)/linux/tpir -ldflags "-X main.version=$(VERSION)" cmd/tpir/main.go

build_macos: setup
	@env GOOS=darwin GOARCH=arm64 go build -o $(OUTPUT)/macos/tpir -ldflags "-X main.version=$(VERSION)" cmd/tpir/main.go

pkg_linux: build_linux
	@tar -C $(OUTPUT)/linux -czf $(OUTPUT)/tpir$(VERSION).linux-amd64.tar.gz tpir

pkg_macos: build_macos
	@tar -C $(OUTPUT)/macos -czf $(OUTPUT)/tpir$(VERSION).macos-arm64.tar.gz tpir

all: clean pkg_linux pkg_macos
