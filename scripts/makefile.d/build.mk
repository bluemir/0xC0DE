##@ Build
## Go Sources
GO_SOURCES = $(shell find . -name "vendor"  -prune -o \
                            -type f -name "*.go" -print)

build/docker-image: $(GO_SOURCES)

# dev build (default, serves source files directly)
# no tag means //go:build !prod is used automatically
.PHONY: build
build: build/$(APP_NAME) ## Build web app (dev mode)

.PHONY: test
test: fmt vet | runtime/tools/go ## Run test
	go test -trimpath ./...

build/$(APP_NAME): $(GO_SOURCES) $(MAKEFILE_LIST) | fmt vet test runtime/tools/go
	@mkdir -p build
	go build -v  \
		-trimpath \
		-ldflags "\
			-X '$(IMPORT_PATH)/internal/buildinfo.AppName=$(APP_NAME)' \
			-X '$(IMPORT_PATH)/internal/buildinfo.Version=$(VERSION)' \
			-X '$(IMPORT_PATH)/internal/buildinfo.BuildTime=$(shell go run scripts/tools/date/main.go)' \
		" \
		$(OPTIONAL_BUILD_ARGS) \
		-o $@ .

# prod build (for deployment, minified + embedded)
# -tags prod enables //go:build prod files
.PHONY: prod
prod: build/$(APP_NAME)-$(VERSION) ## Build web app (prod mode, minified + embedded)

build/$(APP_NAME)-$(VERSION): $(GO_SOURCES) $(MAKEFILE_LIST) | fmt vet gen test runtime/tools/go
	@mkdir -p build
	go build -v \
		-tags prod \
		-trimpath \
		-ldflags "\
			-X '$(IMPORT_PATH)/internal/buildinfo.AppName=$(APP_NAME)' \
			-X '$(IMPORT_PATH)/internal/buildinfo.Version=$(VERSION)' \
			-X '$(IMPORT_PATH)/internal/buildinfo.BuildTime=$(shell go run scripts/tools/date/main.go)' \
		" \
		$(OPTIONAL_BUILD_ARGS) \
		-o $@ .


.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: vulncheck
vulncheck: runtime/tools/govulncheck ## Run govulncheck
	govulncheck ./...

.PHONY: sec
sec: runtime/tools/gosec ## Run gosec
	./runtime/tools/gosec -quiet ./...

.PHONY: gen
gen: runtime/tools/esbuild ## Run go generate
	PATH=$(shell pwd)/runtime/tools:$(PATH) go generate -tags prod -x ./...


runtime/tools/go:
	@which $(notdir $@) || echo "see https://golang.org/doc/install"

runtime/tools/govulncheck:
	GOBIN=$(shell pwd)/runtime/tools go install golang.org/x/vuln/cmd/govulncheck@latest

runtime/tools/gosec:
	GOBIN=$(shell pwd)/runtime/tools go install github.com/securego/gosec/v2/cmd/gosec@latest

build-tools: runtime/tools/govulncheck runtime/tools/gosec runtime/tools/go
