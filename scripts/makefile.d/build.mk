##@ Build
## Go Sources
GO_SOURCES = $(shell find . -name "vendor"  -prune -o \
                            -type f -name "*.go" -print)

build/docker-image: $(GO_SOURCES)

.PHONY: build
build: build/$(APP_NAME) ## Build web app

.PHONY: test
test: fmt vet ## Run test
	@$(MAKE) build/tools/go
	go test -trimpath ./...

build/$(APP_NAME): $(GO_SOURCES) $(MAKEFILE_LIST) fmt vet gen
	@$(MAKE) build/tools/go
	@mkdir -p build
	go build -v \
		-trimpath \
		-ldflags "\
			-X '$(IMPORT_PATH)/internal/buildinfo.AppName=$(APP_NAME)' \
			-X '$(IMPORT_PATH)/internal/buildinfo.Version=$(VERSION)' \
			-X '$(IMPORT_PATH)/internal/buildinfo.BuildTime=$(shell go run scripts/tools/date/main.go)' \
		" \
		$(OPTIONAL_BUILD_ARGS) \
		-o $@ .

build-tools: build/tools/go
build/tools/go:
	@which $(notdir $@) || echo "see https://golang.org/doc/install"

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...
gen: ## Run go generate
	go generate -x ./...
