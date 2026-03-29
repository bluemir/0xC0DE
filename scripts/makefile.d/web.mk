##@ Web

## FE sources
JS_SOURCES    := $(shell find assets/src/js         -type f -name '*.js' ! -name 'index.js' -print -o \
                                                    -type f -name '*.jsx'  -print -o \
                                                    -type f -name '*.json' -print)
CSS_SOURCES   := $(shell find assets/src/css        -type f -name '*.css'  -print)
WEB_LIBS      := $(shell find assets/vendor         -type f                -print)
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

build/$(APP_NAME):            package.json node_modules/.package-lock.json
build/$(APP_NAME)-$(VERSION): package.json node_modules/.package-lock.json


node_modules/.package-lock.json: package.json package-lock.json | runtime/tools/npm
	@mkdir -p $(dir $@)
	npm install

runtime/tools/npm:
	@which $(notdir $@)

## Go esbuild (go tool 사용 - go.mod tool 디렉티브로 버전 고정)
build-tools: runtime/tools/npm

## prod build uses dist folder
OPTIONAL_CLEAN += assets/dist

## vendor → bundle 빌드 (외부 라이브러리)
OPTIONAL_CLEAN += assets/bundle

assets/bundle/bm.js/bm.module.js: assets/vendor/bm.module.js assets/vendor/bm.js/*
	@mkdir -p $(dir $@)
	go tool esbuild $< --bundle --format=esm --outfile=$@

assets/bundle/lit-html/lit-html.js: assets/vendor/lit-html.js package.json package-lock.json | runtime/tools/npm
	@mkdir -p $(dir $@)
	go tool esbuild $< --bundle --format=esm --outfile=$@

assets/bundle/fonts/fonts.css: assets/vendor/fonts.css package.json package-lock.json | runtime/tools/npm
	@mkdir -p $(dir $@)
	go tool esbuild $< --bundle --outdir=assets/bundle/fonts --loader:.woff2=file --asset-names=[name]

build/$(APP_NAME):            assets/bundle/bm.js/bm.module.js assets/bundle/lit-html/lit-html.js assets/bundle/fonts/fonts.css
build/$(APP_NAME)-$(VERSION): assets/bundle/bm.js/bm.module.js assets/bundle/lit-html/lit-html.js assets/bundle/fonts/fonts.css
