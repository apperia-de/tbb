# See https://makefiletutorial.com for in depth examples

CURRENT_VERSION := $(shell cat VERSION)
NEXT_MAJOR_VERSION := $(shell sh scripts/semverinc.sh $(CURRENT_VERSION) 0)
NEXT_MINOR_VERSION := $(shell sh scripts/semverinc.sh $(CURRENT_VERSION) 1)
NEXT_PATCH_VERSION := $(shell sh scripts/semverinc.sh $(CURRENT_VERSION) 2)

.DEFAULT_GOAL := help

.PHONY: help
help: ## Display this help screen
	@echo "Available Makefile targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}'

.PHONY: build-timezone-data
build-timezone-data: ## Build and generate timezone data assets
	mkdir -p bin
	go build -o bin/tz cmd/timezone/main.go
	./bin/tz -build -db assets/timezone.data

.PHONY: test
test: ## Run package unit tests
	go test ./...

.PHONY: test-verbose
test-verbose: ## Run unit tests with verbose logging
	go test -v ./...

.PHONY: test-race
test-race: ## Run unit tests with race detector enabled
	go test -race ./...

.PHONY: code-coverage
code-coverage: ## Run unit tests and show coverage report
	go test -cover -coverprofile=coverage.out ./...
	gocovsh

.PHONY: fmt
fmt: ## Run go fmt on all Go source files
	go fmt ./...

.PHONY: vet
vet: ## Run go vet static analysis
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint static check
	golangci-lint run

.PHONY: build-example
build-example: ## Compile the cmd/example application
	mkdir -p bin
	go build -o bin/example cmd/example/main.go

.PHONY: test-versioning
test-versioning: ## Print potential next semantic versions
	@echo Sematic versioning test...
	@echo ___Current version: v$(CURRENT_VERSION)
	@echo Next patch version: v$(NEXT_PATCH_VERSION)
	@echo Next minor version: v$(NEXT_MINOR_VERSION)
	@echo Next major version: v$(NEXT_MAJOR_VERSION)

.PHONY: update
update: update-deps fmt vet lint test-verbose ## Run all dependency updates, formats, lints and tests

.PHONY: update-deps
update-deps: ## Update go dependencies to their latest minor versions
	go get -u ./...
	go mod tidy

.PHONY: create-next-major-version
create-next-major-version: ## Tag and release next major version
	@echo    Current Version: v$(CURRENT_VERSION)
	@echo Next Major Version: v$(NEXT_MAJOR_VERSION)
	@echo $(NEXT_MAJOR_VERSION) > VERSION
	@git add .
	@git commit -m "Release new major version: (v$(NEXT_MAJOR_VERSION))"
	@git tag v$(NEXT_MAJOR_VERSION)
	@echo In order to update tags run: git push --tags

.PHONY: create-next-minor-version
create-next-minor-version: ## Tag and release next minor version
	@echo    Current Version: v$(CURRENT_VERSION)
	@echo Next Minor Version: v$(NEXT_MINOR_VERSION)
	@echo $(NEXT_MINOR_VERSION) > VERSION
	@git add .
	@git commit -m "Release new minor version: (v$(NEXT_MINOR_VERSION))"
	@git tag v$(NEXT_MINOR_VERSION)
	@echo In order to update tags run: git push --tags

.PHONY: create-next-patch-version
create-next-patch-version: ## Tag and release next patch version
	@echo    Current Version: v$(CURRENT_VERSION)
	@echo Next Patch Version: v$(NEXT_PATCH_VERSION)
	@echo $(NEXT_PATCH_VERSION) > VERSION
	@git add .
	@git commit -m "Release new patch version: (v$(NEXT_PATCH_VERSION))"
	@git tag v$(NEXT_PATCH_VERSION)
	@echo In order to update tags run: git push --tags