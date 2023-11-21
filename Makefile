NAME                                           := ocm
REPO_ROOT                                      := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
GITHUBORG                                      ?= mandelsoft
VERSION                                        := $(shell cat $(REPO_ROOT)/VERSION)
EFFECTIVE_VERSION                              := $(VERSION)+$(shell git rev-parse HEAD)
GIT_TREE_STATE                                 := $(shell [ -z "$$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)
COMMIT                                         := $(shell git rev-parse --verify HEAD)

SOURCES := $(shell go list -f '{{$$I:=.Dir}}{{range .GoFiles }}{{$$I}}/{{.}} {{end}}' ./... )
GOPATH                                         := $(shell go env GOPATH)

NOW         := $(shell date -u +%FT%T%z)
BUILD_FLAGS := "-s -w \
 -X github.com/mandelsoft/mdref/version.gitVersion=$(EFFECTIVE_VERSION) \
 -X github.com/mandelsoft/mdref/version.gitTreeState=$(GIT_TREE_STATE) \
 -X github.com/mandelsoft/mdref/version.gitCommit=$(COMMIT) \
 -X github.com/mandelsoft/mdref/version.buildDate=$(NOW)"

build: ${SOURCES}
	mkdir -p bin
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/mdref .
	bin/mdref src .
