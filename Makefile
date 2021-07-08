VERSION?=$(shell git describe --tags --dirty --always)
export VERSION

IMPORT_PATH=$(shell cat go.mod | head -n 1 | awk '{print $$2}')
APP_NAME=$(notdir $(IMPORT_PATH))

export GO111MODULE=on

# go build args
OPTIONAL_BUILD_ARGS :=

default: build

# sub-makefiles
# for build tools, docker build, deploy, static web files.
include scripts/makefile.d/*

build: build/$(APP_NAME)

run: build/$(APP_NAME)
	$< -vvv server

test:
	go test -v ./pkg/... ./internal/...

clean:
	rm -rf build/ $(OPTIONAL_CLEAN_DIR)

tools: build-tools
	@echo "--- done ---"

help: CMD_LIST=$(shell echo $(MAKEFILE_LIST) | xargs grep -h  "^.PHONY" | awk -F": " '{print $$2}' | tr ' ' '\n' | sort | tr '\n' ' ')
help:
	# requirement
	#  - golang: 1.16.x
	#  - node  : 14.16.x
	#  - make  : 4.3
	#
	# available command :
	#    $(CMD_LIST)

.PHONY: default build run test clean tools build-tools help

