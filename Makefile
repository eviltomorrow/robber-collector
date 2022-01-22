# This how we want to name the binary output
#
version=v1.0

GOPATH ?= $(shell go env GOPATH)
# Ensure GOPATH is set before running build process.
ifeq "$(GOPATH)" ""
  $(error Please set the environment variable GOPATH before running `make`)
endif
PATH := ${GOPATH}/bin:$(PATH)
GCFLAGS=-gcflags "all=-trimpath=${GOPATH}"

GITTAG := $(shell git describe --tags --always)
GITSHA := $(shell git rev-parse --short HEAD)
GITBRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BUILDTIME=`date +%FT%T%z`

LDFLAGS=-ldflags "-X main.GitSha=${GITSHA} -X main.GitTag=${GITTAG} -X main.GitBranch=${GITBRANCH} -X main.BuildTime=${BUILDTIME} -s -w"

# colors compatible setting
CRED:=$(shell tput setaf 1 2>/dev/null)
CGREEN:=$(shell tput setaf 2 2>/dev/null)
CYELLOW:=$(shell tput setaf 3 2>/dev/null)
CEND:=$(shell tput sgr0 2>/dev/null)

.PHONY: all
all: | fmt build

.PHONY: go_version_check
GO_VERSION_MIN=1.17
# Parse out the x.y or x.y.z version and output a single value x*10000+y*100+z (e.g., 1.9 is 10900)
# that allows the three components to be checked in a single comparison.
VER_TO_INT:=awk '{split(substr($$0, match ($$0, /[0-9\.]+/)), a, "."); print a[1]*10000+a[2]*100+a[3]}'
go_version_check:
	@echo "$(CGREEN)=> Go version check ...$(CEND)"
	@if test $(shell go version | $(VER_TO_INT) ) -lt \
  	$(shell echo "$(GO_VERSION_MIN)" | $(VER_TO_INT)); \
  	then printf "go version $(GO_VERSION_MIN)+ required, found: "; go version; exit 1; \
		else echo "go version check pass";	fi

# Code format
.PHONY: fmt
fmt: go_version_check
	@echo "$(CGREEN)=> Run gofmt on all source files ...$(CEND)"
	@echo "gofmt -l -s -w ..."
	@ret=0 && for d in $$(go list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		gofmt -l -s -w $$d/*.go || ret=$$? ; \
	done ; exit $$ret

# Run golang test cases
.PHONY: test
test:
	@echo "$(CGREEN)=> Run all test cases ...$(CEND)"
	@go test $(LDFLAGS) -timeout 10m -race ./...
	@echo "$(CGREEN)=> Test Success!$(CEND)"

# Compile
.PHONY: compile	
compile:
	@echo "$(CGREEN)=> Compile protobuf ...$(CEND)"
	@bash scripts/compile.sh


# Builds the project
.PHONY: build
build: fmt
	@mkdir -p examples/bin
	@mkdir -p examples/etc
	@cp configs/config.toml examples/etc/config.toml
	@echo "$(CGREEN)=> Building binary...$(CEND)"
	go build -race ${LDFLAGS} ${GCFLAGS} -o examples/bin/robber-collector.go cmd/main.go
	@echo "$(CGREEN)=> Build Success!$(CEND)"


