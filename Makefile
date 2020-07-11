GO ?= $(shell command -v go 2> /dev/null)
GOPATH ?= $(shell go env GOPATH)

export GO111MODULE=on

.PHONY: default
default: update dist

.PHONY: dist
dist: update
	rm -rf dist/
	mkdir -p dist/

	go install github.com/markbates/pkger/cmd/pkger
	$(GOPATH)/bin/pkger

	# "-tags 'osusergo netgo'" is needed for creating static binary.
	# refs: https://github.com/golang/go/issues/26492#issuecomment-635563222
	env GOOS=linux   GOARCH=amd64 $(GO) build -tags 'osusergo netgo' -o dist/dio-exporter-linux-amd64       github.com/kaakaa/dio-exporter/cmd/dio-exporter;
	env GOOS=darwin  GOARCH=amd64 $(GO) build -tags 'osusergo netgo' -o dist/dio-exporter-darwin-amd64      github.com/kaakaa/dio-exporter/cmd/dio-exporter;
	env GOOS=windows GOARCH=amd64 $(GO) build -tags 'osusergo netgo' -o dist/dio-exporter-windows-amd64.exe github.com/kaakaa/dio-exporter/cmd/dio-exporter;

.PHONY: update
update:
	git submodule update --init

test: dist
	go test -v ./test

debug-server: update
	go run cmd/debug-server/main.go
