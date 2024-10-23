##@ Web
## FE sources
JS_SOURCES    := $(shell find assets/js             -type f -name '*.js'   -print -o \
                                                    -type f -name '*.jsx'  -print -o \
                                                    -type f -name '*.json' -print)
CSS_SOURCES   := $(shell find assets/css            -type f -name '*.css'  -print)
WEB_LIBS      := $(shell find assets/lib            -type f                -print)
HTML_SOURCES  := $(shell find assets/html-templates -type f -name '*.html' -print)
IMAGES        := $(shell find assets/images         -type f                -print)
WEB_META      := assets/manifest.json assets/favicon.ico

build/docker-image: $(JS_SOURCES) $(CSS_SOURCES) $(WEB_LIBS) $(HTML_SOURCES)

.PHONY: web
web: ## Build web-files. (bundle, minify, transpile, etc.)

## common static files
web: $(WEB_LIBS:assets/%=build/static/%)
web: $(IMAGES:assets/%=build/static/%)
web: $(WEB_META:assets/%=build/static/%)

build/static/%: assets/%
	@mkdir -p $(dir $@)
	cp $< $@

## js import helper
build/static/js/index.js: assets/js/index.js
assets/js/index.js: $(JS_SOURCES) scripts/tools/import-helper/*
	go run ./scripts/tools/import-helper \
		--dir=assets/js \
		--target=$@
OPTIONAL_CLEAN += assets/js/index.js

## esbuild
web: build/static/js/index.js  # entrypoint
build/static/js/%: export NODE_PATH=assets/js:assets/lib
build/static/js/%: assets/js/% package-lock.json $(MAKEFILES)
	@$(MAKE) build/tools/npx
	@mkdir -p $(dir $@)
	npx esbuild $< --outdir=$(dir $@) \
		--bundle --sourcemap --format=esm --minify \
		--external:lit-html \
		$(OPTIONAL_WEB_BUILD_ARGS)

web: build/static/css/page.css build/static/css/element.css
build/static/css/%: assets/css/%
	@$(MAKE) build/tools/npx
	@mkdir -p $(dir $@)
	npx esbuild $< --outdir=$(dir $@) \
		--bundle --sourcemap --minify \
		$(OPTIONAL_WEB_BUILD_ARGS)

build/static/css/page.css build/static/css/element.css: $(CSS_SOURCES)

build/$(APP_NAME): web $(HTML_SOURCES)

## resolve depandancy
OPTIONAL_CLEAN += node_modules

build/$(APP_NAME): package-lock.json
build/docker-image: package-lock.json

package-lock.json: package.json
	@$(MAKE) build/tools/npm
	@mkdir -p $(dir $@)
	npm install
yarn.lock: package.json
	@$(MAKE) build/tools/yarn
	@mkdir -p $(dir $@)
	yarn install

build-tools: build/tools/npm build/tools/yarn build/tools/npx
build/tools/npm:
	@which $(notdir $@)
build/tools/npx:
	@which $(notdir $@)
build/tools/yarn: build/tools/npm
	@which $(notdir $@) || (npm install -g yarn)
