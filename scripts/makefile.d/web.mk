##@ Web

## FE sources
JS_SOURCES    := $(shell find assets/src/js         -type f -name '*.js'   -print -o \
                                                    -type f -name '*.jsx'  -print -o \
                                                    -type f -name '*.json' -print)
CSS_SOURCES   := $(shell find assets/src/css        -type f -name '*.css'  -print)
WEB_LIBS      := $(shell find assets/lib            -type f                -print)
HTML_SOURCES  := $(shell find assets/html-templates -type f -name '*.html' -print)
IMAGES        := $(shell find assets/images         -type f                -print)
WEB_META      := assets/manifest.json assets/favicon.ico

.PHONY: web
web: ## Build web-files. (bundle, minify, transpile, etc.)

## common static files
web: $(WEB_LIBS) $(IMAGES) $(WEB_META)
build/$(APP_NAME):  $(WEB_LIBS) $(IMAGES) $(WEB_META)

## js import helper
OPTIONAL_CLEAN += assets/src/js/index.js
assets/src/js/index.js: $(JS_SOURCES) scripts/tools/import-helper/*
	mkdir -p $(dir $@)
	go run ./scripts/tools/import-helper --dir=assets/src/js --target=$@

## js build, with esbuild
web: assets/js/index.js # entrypoints
build/$(APP_NAME): assets/js/index.js

assets/js/%: export NODE_PATH=assets/src/js:assets/lib
assets/js/%: assets/src/js/% package.json package-lock.json | build/tools/npx
	@mkdir -p $(dir $@)
	npx esbuild $< --outdir=$(dir $@) \
		--bundle --sourcemap --format=esm \
		--external:lit-html \
		--external:bm.js/bm.module.js \
		$(OPTIONAL_WEB_BUILD_ARGS)
OPTIONAL_CLEAN += assets/js

## css build, with esbuild
web: assets/css/page.css assets/css/element.css
build/$(APP_NAME): assets/css/page.css assets/css/element.css
assets/css/%: assets/src/css/% | build/tools/npx
	@mkdir -p $(dir $@)
	npx esbuild $< --outdir=$(dir $@) \
		--bundle --sourcemap \
		$(OPTIONAL_WEB_BUILD_ARGS)
OPTIONAL_CLEAN += assets/css

assets/css/page.css assets/css/element.css: $(CSS_SOURCES) # TODO: import graph?

# html
build/$(APP_NAME): $(HTML_SOURCES)
build/docker-image: $(JS_SOURCES) $(CSS_SOURCES) $(WEB_LIBS) $(HTML_SOURCES)

## resolve depandancy
OPTIONAL_CLEAN += node_modules

package-lock.json: package.json | build/tools/npm
	@mkdir -p $(dir $@)
	npm install

build-tools: build/tools/npm build/tools/npx
build/tools/npm:
	@which $(notdir $@)
build/tools/npx:
	@which $(notdir $@)
build/tools/yarn: |  build/tools/npm
	@which $(notdir $@) || (npm install -g yarn)

vet: assets/js/.placeholder assets/css/.placeholder
