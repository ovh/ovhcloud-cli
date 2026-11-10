GIT_DESCRIBE ?= $(shell git describe --tags 2>/dev/null)
SNAPSHOT	 := $(shell svu next 2>/dev/null)
VERSION		 ?= $(if ${GIT_TAG},${GIT_DESCRIBE},${SNAPSHOT}-snapshot${CDS_RUN_NUMBER})

LD_PKG	  = $(shell go list ./internal/version)
LD_FLAGS  = -s -w -extldflags -static -X ${LD_PKG}.Version=${VERSION}
BUILD_CMD = CGO_ENABLED=0 go build

all: build

build:
	${BUILD_CMD} -trimpath -ldflags "${LD_FLAGS}" -o ovhcloud ./cmd/ovhcloud

wasm:
	GOOS=js GOARCH=wasm ${BUILD_CMD} -trimpath -ldflags "${LD_FLAGS}" -o ovhcloud.wasm ./cmd/ovhcloud

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

schemas:
	@if [ -z "$(UNIVERSE)" ]; then echo "Usage: make schemas UNIVERSE=<name> (e.g. cloud, domain, vps)"; exit 1; fi
	@tmp=$$(mktemp internal/assets/api-schemas/$(UNIVERSE).json.XXXXXX) && \
	curl -s "https://eu.api.ovh.com/v1/$(UNIVERSE).json?format=openapi3" | jq 'del(.paths[] | .[]["x-code-samples"])' > "$$tmp" && \
	mv "$$tmp" internal/assets/api-schemas/$(UNIVERSE).json

# Keeps public routes only; "long" isn't a valid OpenAPI type and breaks the request validator.
SCHEMAS_V2_FILTER = .paths |= with_entries( \
		.value |= with_entries(select(.value | type != "object" or ((.["x-badges"] // []) | any(.label == "Internal use only") | not))) \
		| select(.value | any(type == "object"))) \
	| del(.paths[][] | objects | .["x-code-samples"]) \
	| walk(if type == "object" and .type == "long" then .type = "integer" | .format = "int64" else . end)

schemas-v2:
	@tmp=$$(mktemp internal/assets/api-schemas/cloud_v2.json.XXXXXX) && \
	curl -sf "https://eu.api.ovh.com/v2/publicCloud.json?format=openapi3" | jq '$(SCHEMAS_V2_FILTER)' > "$$tmp" && \
	mv "$$tmp" internal/assets/api-schemas/cloud_v2.json || { rm -f "$$tmp"; exit 1; }

setup:
	curl --proto '=https' --tlsv1.2 -LsSf https://github.com/j178/prek/releases/latest/download/prek-installer.sh | sh
	prek install

.PHONY: all wasm doc schemas schemas-v2 setup
