.PHONY: build
.DEFAULT_GOAL := setup

##@ Global

help: ## Prints help.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Setup

deps: ## Install the repository dependencies (linter, database migration, mocks, etc.).
	go install -v github.com/incu6us/goimports-reviser/v3@latest

setup:
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
	go build -tags=go_json -tags appsec -o .bin/ github.com/shortcuts/codes/cmd

clean: ## Cleans the bin folder.
	rm -r .bin/ || true

dev: stop ## Runs the service in watch mode
	@make run &
	@fswatch -or --event=Updated cmd/ | xargs -n1 -I{} make restart || exit 0
	make stop

run: ## Run the service
	go run github.com/shortcuts/codes/cmd

restart: ## Stops then run the service
	make -j1 stop run &

stop: ## Stops the server started in the background
	@killall $* &> /dev/null || exit 0

test: ## Runs the test suites
	go test -timeout 30s -race github.com/shortcuts/codes/cmd/...
