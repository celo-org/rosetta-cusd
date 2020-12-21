GO ?= latest

GITHUB_ORG?=celo-org
GITHUB_REPO?=rosetta

GOLANGCI_VERSION=1.32.2
GOLANGCI_exists := $(shell command -v golangci-lint 2> /dev/null)
GOLANGCI_v_installed := $(shell echo $(shell golangci-lint --version) | cut -d " " -f 4)

COMMIT_SHA=$(shell git rev-parse HEAD)

LICENSE_SCRIPT=addlicense -c "Celo Org" -l "apache" -v

all: 
	go build ./...

fmt: 
	go fmt ./...

install-lint-ci:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v$(GOLANGCI_VERSION)

lint: ## Run linters.
ifeq ("$(GOLANGCI_exists)","")
	$(error "No golangci in PATH, consult https://github.com/golangci/golangci-lint#install")
else ifneq ($(GOLANGCI_v_installed), $(GOLANGCI_VERSION))
	$(error "Installed golangci version $(GOLANGCI_v_installed) \
	 does not match expected version $(GOLANGCI_VERSION)")
else
	golangci-lint run -c .golangci.yml
endif

clean:
	go clean -cache

add-license:
	${LICENSE_SCRIPT} services configuration main.go

check-license:
	${LICENSE_SCRIPT} -check services configuration main.go
