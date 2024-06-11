.PHONY: build
.DEFAULT_GOAL := setup

##@ Global

help: ## Prints help.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Setup

deps: ## Install the repository dependencies (linter, database migration, mocks, etc.).
	go install github.com/incu6us/goimports-reviser/v3@latest
	go install github.com/mitranim/gow@latest

setup: ## Cleans and install deps
	make clean
	make deps

##@ Linters

lint: ## Lint Go files
	@goimports-reviser -use-cache \
		-company-prefixes github.com/shortcuts \
		-project-name github.com/shortcuts/codes \
		./...
	golangci-lint run --fix

##@ Development

build: ## Builds the service
	make clean
	CGO_ENABLED=0 GOOS=linux go build -tags=go_json -o .bin/main github.com/shortcuts/codes/cmd

bundle: ## Builds the docker image
	docker build \
	  --build-arg version=$$(git rev-parse HEAD) \
	  -t ghcr.io/shortcuts/codes:latest \
	  -t ghcr.io/shortcuts/codes:$$(git rev-parse HEAD) \
	  -f Dockerfile .


clean: ## Cleans the bin folder.
	rm -r .bin/ || true

dev: stop ## Runs the service in watch mode
	gow -e=go,mod,html,css,md run cmd/main.go

stop: ## Stops leftover services
	kill $$(lsof -t -i:42069) || true

test: ## Runs the test suites
	gow test -timeout 30s -race github.com/shortcuts/codes/cmd/...
