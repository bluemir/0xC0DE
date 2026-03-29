VERSION?=$(shell git describe --tags --dirty --always)
export VERSION

IMPORT_PATH=$(shell cat go.mod | head -n 1 | awk '{print $$2}')
APP_NAME=$(notdir $(IMPORT_PATH))

export GO111MODULE=on
export GOPRIVATE=
export PATH:=./runtime/tools:$(PATH)
export GOTOOLCHAIN=go1.26.0+auto

# go build args
OPTIONAL_BUILD_ARGS?=

ifneq ($(shell printf '%s\n' "$(MIN_MAKE_VERSION)" "$(MAKE_VERSION)" | sort -V | tail -n 1),$(MAKE_VERSION))
    $(error Makefile을 실행하려면 Make 버전 4.3 이상이 필요합니다. 현재 버전: $(MAKE_VERSION))
endif

.PHONY: default
default: | runtime/tools/go
	@go run -C scripts/tools/make-select .

# sub-makefiles
# for build tools, docker build, deploy, static web files.
include scripts/makefile.d/*.mk

##@ General
.PHONY: clean
clean: ## Clean up
	rm -rf build/ $(OPTIONAL_CLEAN)

.PHONY: build-tools
build-tools: ## Install build tools
	# Build tool installed
.PHONY: tools
tools: build-tools ## Install tools(include build tools)
	# Tool installed

.PHONY: help
help: | runtime/tools/go
help: ## Display this help
	@go run -C scripts/tools/make-select . --print-only

%/.placeholder:
	@mkdir -p $(dir $@)
	touch $@
