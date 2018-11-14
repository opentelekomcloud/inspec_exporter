GO           := GO15VENDOREXPERIMENT=1 go
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
PROMU        := $(FIRST_GOPATH)/bin/promu
DEP          := $(FIRST_GOPATH)/bin/dep
pkgs          = $(shell $(GO) list ./... | grep -v /vendor/)

PREFIX                  ?= $(shell pwd)
BIN_DIR                 ?= $(shell pwd)
DOCKER_IMAGE_NAME       ?= inspec_exporter
DOCKER_IMAGE_TAG        ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))


all: format build docker

style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

update: dep
	@echo ">> update vendors"
	@$(DEP) ensure

build: promu
	@echo ">> building binaries"
	@$(PROMU) build --prefix $(PREFIX)

crossbuild: dep promu
	@echo ">> crossbuild binaries"
	@$(PROMU) crossbuild
	@$(PROMU) crossbuild tarballs
	@$(PROMU) checksum .tarballs
	@$(PROMU) release .tarballs

docker: build
	@echo ">> building docker image"
	@docker build -t "$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

promu:
	@GOOS=$(shell uname -s | tr A-Z a-z) \
	GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
	$(GO) get -u github.com/prometheus/promu

dep:
	@GOOS=$(shell uname -s | tr A-Z a-z) \
	GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
	$(GO) get -u github.com/golang/dep/cmd/dep

.PHONY: all style format update build crossbuild docker promu dep