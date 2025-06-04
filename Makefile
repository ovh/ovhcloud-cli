GIT_DESCRIBE ?= $(shell git describe --tags 2>/dev/null || echo "")
LAST_COMMIT	 = $(shell git rev-parse HEAD)
SNAPSHOT	 := $(shell svu next 2>/dev/null || echo "")
VERSION		 ?= $(if ${GIT_TAG},${GIT_DESCRIBE},${SNAPSHOT}-snapshot${CDS_RUN_NUMBER})

LD_PKG	  = stash.ovh.net/api/ovh-cli/internal/cmd
LD_FLAGS  = -s -w -extldflags -static -X ${LD_PKG}.lastCommit=${LAST_COMMIT} -X ${LD_PKG}.version=${VERSION}
BUILD_CMD = CGO_ENABLED=0 go build

all:
	${BUILD_CMD} -ldflags "${LD_FLAGS}" -o ovh-cli .

.PHONY: all