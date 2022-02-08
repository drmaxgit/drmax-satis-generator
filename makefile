PROJECT_NAME := "drmax-satis-generator"
PKG := "github.com/drmaxgit/drmax-satis-generator"
GOARCH := amd64
CI_COMMIT_TAG ?= v0.0.0

.PHONY: all dep clean build

all: clean dep build
test: vet fmt lint

vet: ## Vet code
	@go vet

fmt: ## Format code
	@go fmt

lint: ## Lint code
	@go install golang.org/x/lint/golint
	@golint .

dep: ## Get the dependencies
	@go get

build: dep ## Build the binary file
	export GOARCH=amd64
	GOOS=darwin go build -o bin/$(PROJECT_NAME)-darwin-amd64-$(CI_COMMIT_TAG)
	GOOS=windows go build -o bin/$(PROJECT_NAME)-windows-amd64-$(CI_COMMIT_TAG).exe
	GOOS=linux go build -o bin/$(PROJECT_NAME)-linux-amd64-$(CI_COMMIT_TAG)

clean: ## Remove previous build
	@rm -f bin/$(PROJECT_NAME)-*

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
