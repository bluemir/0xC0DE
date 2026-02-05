##@ Web

## FE sources
JS_SOURCES    := $(shell find assets/src/js         -type f -name '*.js' ! -name 'index.js' -print -o \
                                                    -type f -name '*.jsx'  -print -o \
                                                    -type f -name '*.json' -print)
CSS_SOURCES   := $(shell find assets/src/css        -type f -name '*.css'  -print)
WEB_LIBS      := $(shell find assets/lib            -type f                -print)
HTML_SOURCES  := $(shell find assets/html-templates -type f -name '*.html' -print)
IMAGES        := $(shell find assets/images         -type f                -print)
WEB_META      := assets/manifest.json assets/favicon.ico

## js import helper
OPTIONAL_CLEAN += assets/src/js/index.js
assets/src/js/index.js: $(JS_SOURCES) scripts/tools/import-helper/*
	mkdir -p $(dir $@)
	go run ./scripts/tools/import-helper --dir=assets/src/js --target=$@

## dev build:
build/$(APP_NAME): assets/src/js/index.js

## prod build: esbuild runs via go:generate in assets/static_prod.go

## resolve dependency
OPTIONAL_CLEAN += node_modules

build/$(APP_NAME):            package.json package-lock.json
build/$(APP_NAME)-$(VERSION): package.json package-lock.json

package-lock.json: package.json | runtime/tools/npm
	@mkdir -p $(dir $@)
	npm install

runtime/tools/npm:
	@which $(notdir $@)

## Go esbuild (npm/npx 대신 사용)
runtime/tools/esbuild:
	GOBIN=$(shell pwd)/runtime/tools go install github.com/evanw/esbuild/cmd/esbuild@latest

build-tools: runtime/tools/esbuild runtime/tools/npm

## prod build uses dist folder
OPTIONAL_CLEAN += assets/dist
