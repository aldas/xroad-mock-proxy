PROJECT_NAME := xroad-mock-proxy
PROJECT_EXECUTABLE := xroad-mock-proxy
PKG := "github.com/aldas/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/...)
VERSION := $(shell git describe --always --long --dirty)
BUILD := $(shell date +%FT%T%z)
LDFLAGS=-X main.version=${VERSION} -X main.build=$(BUILD)

.PHONY: all init build clean lint test coverage coverhtml

all: build

init:
	git config core.hooksPath ./scripts/.githooks
	@go get -u golang.org/x/lint/golint

build: ## Build the binary file
	@go build -ldflags "$(LDFLAGS)" -o $(PROJECT_EXECUTABLE) -v $(PKG)

build-for-docker: ## Build the binary file suitable for Docker
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s $(LDFLAGS)" -o $(PROJECT_EXECUTABLE) -v $(PKG)

docker: ## - Build docker image
	@docker build -f Dockerfile -t $(PROJECT_EXECUTABLE) .

run-docker:	## - Run docker image
	@docker run $(PROJECT_EXECUTABLE):latest

clean: ## Remove previous build
	@rm -f $(PROJECT_EXECUTABLE)

lint: ## Lint the files
	@golint -set_exit_status ${PKG_LIST}

vet: ## Vet the files
	@go vet ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}

check: lint vet test ## check project

race: ## Run data race detector
	@go test -race -short ${PKG_LIST}

msan: ## Run memory sanitizer
	@go test -msan -short ${PKG_LIST}

coverage: ## Generate global code coverage report
	bash ./scripts/coverage.sh;

coverhtml: ## Generate global code coverage report in HTML
	bash ./scripts/coverage.sh html;

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
