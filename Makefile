GO ?= $(shell command -v go 2> /dev/null)
GOPATH ?= $(shell go env GOPATH)

export GO111MODULE=on

.PHONY: default
default: update dist

.PHONY: dist
dist: update
	rm -rf dist/
	mkdir -p dist/
	env GOOS=linux GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o dist/dio-exporter-linux-amd64;
	env GOOS=darwin GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o dist/dio-exporter-darwin-amd64;
	env GOOS=windows GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o dist/dio-exporter-windows-amd64.exe;

.PHONY: update
update:
	git submodule update --init
	go install github.com/markbates/pkger/cmd/pkger
	$(GOPATH)/bin/pkger
