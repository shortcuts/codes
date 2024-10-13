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
	prettier --write .

##@ Development

build: ## Builds the service
	cd cmd && ko build . --bare --platform=linux/arm64

clean: ## Cleans the bin folder.
	rm -r .bin/ || true
	docker container stop codes || true
	docker container rm codes || true
	docker image rm -f $$(docker image ls ko.local -q) || true

dev: stop ## Runs the service in watch mode
	DEV=true gow -e=go,mod,html,css,md,js run cmd/main.go

start: stop build ## Stops everything, build for prod, starts the image
	docker compose up -d

stop: ## Stops leftover services
	kill $$(lsof -t -i:1313) || true

test: ## Runs the test suites
	gow test -timeout 30s -race github.com/shortcuts/codes/cmd/...
