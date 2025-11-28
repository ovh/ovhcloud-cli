GIT_DESCRIBE ?= $(shell git describe --tags 2>/dev/null)
SNAPSHOT	 := $(shell svu next 2>/dev/null)
VERSION		 ?= $(if ${GIT_TAG},${GIT_DESCRIBE},${SNAPSHOT}-snapshot${CDS_RUN_NUMBER})

LD_PKG	  = $(shell go list ./internal/version)
LD_FLAGS  = -s -w -extldflags -static -X ${LD_PKG}.Version=${VERSION}
BUILD_CMD = CGO_ENABLED=0 go build

all:
	${BUILD_CMD} -ldflags "${LD_FLAGS}" -o ovhcloud ./cmd/ovhcloud

wasm:
	GOOS=js GOARCH=wasm ${BUILD_CMD} -ldflags "${LD_FLAGS}" -o ovhcloud.wasm ./cmd/ovhcloud

test:
	go test -v ./...

fmt:
	go fmt ./...

doc:
	go run cmd/docgen/main.go
	git checkout doc/ovhcloud.md

release-snapshot:
	goreleaser release --snapshot --clean --parallelism 1

release:
	goreleaser release --clean

.PHONY: all wasm doc
