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

SCHEMAS_DIR  = internal/assets/api-schemas
SCHEMAS_ROOT ?= https://eu.api.ovh.com

# fetch-schema downloads one OpenAPI document and installs it only once it has
# been shown to be one. $(1) is the URL, $(2) the file to write under
# $(SCHEMAS_DIR).
#
# The checks are not decoration. Piping curl straight into jq hides a failure:
# an unreachable host or a proxy that answers nothing writes an empty body and
# curl's exit code is lost to the pipe, jq turns nothing into nothing and exits
# 0, and the mv then installs an empty file over a working schema. An empty
# schema embeds, builds, and ships; it surfaces much later, when a command asks
# it for an enumeration and gets "value of openapi must be a non-empty string"
# instead of a list. A 404 happens to be caught today because jq cannot iterate
# over the error document, which is luck rather than a check.
#
# So the document is downloaded to a file, where curl's own exit code is the one
# being tested, and it has to declare an openapi version and at least one path
# before it may replace anything.
define fetch-schema
	@raw=$$(mktemp "$(SCHEMAS_DIR)/$(2).raw.XXXXXX"); \
	out=$$(mktemp "$(SCHEMAS_DIR)/$(2).new.XXXXXX"); \
	trap 'rm -f "$$raw" "$$out"' EXIT INT TERM; \
	curl -fsS "$(1)" -o "$$raw" && \
	jq -e '(.openapi | type) == "string" and (.paths | length) > 0' "$$raw" > /dev/null && \
	jq 'del(.paths[] | .[]["x-code-samples"])' "$$raw" > "$$out" && \
	mv "$$out" "$(SCHEMAS_DIR)/$(2).json" && \
	echo "installed $(SCHEMAS_DIR)/$(2).json ($$(jq '.paths | length' "$(SCHEMAS_DIR)/$(2).json") paths)"
endef

schemas:
	@if [ -z "$(UNIVERSE)" ]; then echo "Usage: make schemas UNIVERSE=<name> (e.g. cloud, domain, vps)"; exit 1; fi
	$(call fetch-schema,$(SCHEMAS_ROOT)/v1/$(UNIVERSE).json?format=openapi3,$(UNIVERSE))

# schemas-v2 fetches from the other catalogue. It is a separate target because
# the two are addressed differently: v1 is one name per universe, v2 is a path,
# and two of them ("dedicated/server", "publicCloud") do not even resemble the
# file they are stored under. NAME therefore has to be given rather than
# derived, which also keeps the existing cloud_v2.json name reachable.
#
# What this target does NOT do is curate. The v2 schemas in this repository are
# a hand-picked subset of the paths the CLI exposes, and choosing what belongs
# in one is a judgement call, not a transformation. It does print what it just
# pulled in, broken down by maturity badge, because "Internal use only" is the
# one thing a curator has to look at and it is invisible in a 2 MB diff.
schemas-v2:
	@if [ -z "$(API)" ] || [ -z "$(NAME)" ]; then \
		echo "Usage: make schemas-v2 API=<path> NAME=<file> (e.g. API=dedicated/server NAME=baremetal_v2)"; \
		exit 1; \
	fi
	$(call fetch-schema,$(SCHEMAS_ROOT)/v2/$(API).json?format=openapi3,$(NAME))
	@jq -r '[.paths[] | .[] | select(type == "object") | ((.["x-badges"] // [{label: "no badge"}]) | .[] | .label)] \
		| group_by(.) | sort_by(-length) | .[] | "  \(length)\t\(.[0])"' "$(SCHEMAS_DIR)/$(NAME).json"

setup:
	curl --proto '=https' --tlsv1.2 -LsSf https://github.com/j178/prek/releases/latest/download/prek-installer.sh | sh
	prek install

.PHONY: all wasm doc schemas schemas-v2 setup
