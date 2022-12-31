# MAKEFILE
#
# @author      Nicola Asuni
# @link        https://github.com/tecnickcom/rpistat
# ------------------------------------------------------------------------------

SHELL=/bin/bash
.SHELLFLAGS=-o pipefail -c

# CVS path (path to the parent dir containing the project)
CVSPATH=github.com/tecnickcom/rpistat

# Project owner
OWNER=tecnickcom

# Project vendor
VENDOR=tecnickcom

# Project name
PROJECT=rpistat

# Project version
VERSION=$(shell cat VERSION)

# Project release number (packaging build number)
RELEASE=$(shell cat RELEASE)

# Current directory
CURRENTDIR=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))

# Target directory
TARGETDIR=$(CURRENTDIR)target

# Directory where to store binary utility tools
BINUTIL=$(TARGETDIR)/binutil

# GO lang path
ifeq ($(GOPATH),)
	# extract the GOPATH
	GOPATH=$(firstword $(subst /src/, ,$(CURRENTDIR)))
endif

# Add the GO binary dir in the PATH
export PATH := $(GOPATH)/bin:$(PATH)

# Path for binary files (where the executable files will be installed)
BINPATH=usr/local/bin/

# STATIC is a flag to indicate whether to build using static or dynamic linking
STATIC=1
ifeq ($(STATIC),0)
	STATIC_TAG=dynamic
	STATIC_FLAG=
else
	STATIC_TAG=static
	STATIC_FLAG=-static
endif

# Common commands
GO=GOPATH=$(GOPATH) GOPRIVATE=$(CVSPATH) go
GOFMT=gofmt
GOTEST=GOPATH=$(GOPATH) gotest
GODOC=GOPATH=$(GOPATH) godoc

# Current operating system and architecture as one string.
GOOSARCH=$(shell go env GOOS GOARCH | tr -d \\n)

# OS and Architecture used to build the Go binary for Docker.
LINUXGOBUILDENV=GOOS=linux GOARCH=amd64

# Environment variables for the go build command
GOBUILDENV=env GOOS=linux GOARCH=arm64
#GOBUILDENV=

# Directory containing the source code
CMDDIR=./cmd

# List of packages
GOPKGS=$(shell $(GO) list $(CMDDIR)/... )

# Enable junit report when not in LOCAL mode
ifeq ($(strip $(DEVMODE)),LOCAL)
	TESTEXTRACMD=&& $(GO) tool cover -func=$(TARGETDIR)/report/coverage.out
else
	TESTEXTRACMD=2>&1 | tee >(PATH=$(GOPATH)/bin:$(PATH) go-junit-report > $(TARGETDIR)/test/report.xml); test $${PIPESTATUS[0]} -eq 0
endif

# Display general help about this command
.PHONY: help
help:
	@echo ""
	@echo "$(PROJECT) Makefile."
	@echo "GOPATH=$(GOPATH)"
	@echo "The following commands are available:"
	@echo ""
	@echo "    make build         : Compile the application"
	@echo "    make clean         : Remove any build artifact"
	@echo "    make deps          : Get dependencies"
	@echo "    make format        : Format the source code"
	@echo "    make linter        : Check code against multiple linters"
	@echo "    make mod           : Download dependencies"
	@echo "    make modupdate     : Update dependencies"
	@echo "    make qa            : Run all tests and static analysis tools"
	@echo ""
	@echo "Use DEVMODE=LOCAL for human friendly output."
	@echo "To test and build everything from scratch:"
	@echo "DEVMODE=LOCAL make format clean mod deps qa build"
	@echo ""

# Alias for help target
all: help

# Compile the application
.PHONY: build
build:
	$(GOBUILDENV) \
	CGO_ENABLED=0 \
	$(GO) build \
	-tags ${STATIC_TAG} \
	-ldflags '-w -s -X main.programVersion=${VERSION} -X main.programRelease=${RELEASE} -extldflags "-fno-PIC ${STATIC_FLAG}"' \
	-o ./target/${BINPATH}$(PROJECT) $(CMDDIR)

# Remove any build artifact
.PHONY: clean
clean:
	rm -rf $(TARGETDIR)

# Get the test dependencies
.PHONY: deps
deps: ensuretarget
	curl --silent --show-error --fail --location https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BINUTIL) v1.50.1
	#$(GO) install github.com/rakyll/gotest
	#$(GO) install github.com/jstemmer/go-junit-report
	#$(GO) install github.com/golang/mock/mockgen

# Create the trget directories if missing
.PHONY: ensuretarget
ensuretarget:
	@mkdir -p $(TARGETDIR)/test
	@mkdir -p $(TARGETDIR)/report
	@mkdir -p $(TARGETDIR)/binutil

# Format the source code
.PHONY: format
format:
	@find $(CMDDIR) -type f -name "*.go" -exec $(GOFMT) -s -w {} \;

# Execute multiple linter tools
.PHONY: linter
linter:
	@echo -e "\n\n>>> START: Static code analysis <<<\n\n"
	$(BINUTIL)/golangci-lint run --exclude-use-default=false $(CMDDIR)/...
	@echo -e "\n\n>>> END: Static code analysis <<<\n\n"

# Download dependencies
.PHONY: mod
mod:
	$(GO) mod download all

# Update dependencies
.PHONY: modupdate
modupdate:
	$(GO) get $(shell $(GO) list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)

# Run all tests and static analysis tools
.PHONY: qa
qa: linter
