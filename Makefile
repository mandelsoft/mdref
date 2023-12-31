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
 -X main.gitVersion=$(EFFECTIVE_VERSION) \
 -X main.gitTreeState=$(GIT_TREE_STATE) \
 -X main.gitCommit=$(COMMIT) \
 -X main.buildDate=$(NOW)"

.PHONY: build
build: bin/mdref
	bin/mdref --list --headings src .

bin/mdref: ${SOURCES}
	mkdir -p bin
	CGO_ENABLED=0 go build -ldflags $(BUILD_FLAGS) -o bin/mdref .

.PHONY: test
test: bin/mdref
	bin/mdref --list --headings src .
	diff -ur doc test/doc
	diff README.md test/README.md

