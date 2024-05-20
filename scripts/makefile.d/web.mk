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

STATICS :=

## common static files
STATICS += $(WEB_LIBS:assets/%=build/static/%)
STATICS += $(IMAGES:assets/%=build/static/%)
STATICS += $(WEB_META:assets/%=build/static/%)

build/static/%: assets/%
	@mkdir -p $(dir $@)
	cp $< $@

## js import helper
build/static/js/index.js: assets/js/index.js
assets/js/index.js: $(JS_SOURCES) scripts/tools/import-helper/*
	go run ./scripts/tools/import-helper > $@
OPTIONAL_CLEAN += assets/js/index.js

## esbuild
STATICS += build/static/js/index.js # entrypoint
build/static/js/%: export NODE_PATH=assets/js:assets/lib
build/static/js/%: $(JS_SOURCES) build/yarn-updated
	@$(MAKE) build/tools/npx
	@mkdir -p $(dir $@)
	npx esbuild $(@:build/static/%=assets/%) --outdir=$(dir $@) \
		--bundle --sourcemap --format=esm \
		--external:lit-html \
		$(OPTIONAL_WEB_BUILD_ARGS)
	#--external:/config.js \
	#--minify \

STATICS += build/static/css/page.css build/static/css/element.css
build/static/css/%: $(CSS_SOURCES)
	@$(MAKE) build/tools/npx
	@mkdir -p $(dir $@)
	npx esbuild $(@:build/static/%=assets/%) --outdir=$(dir $@) \
		--bundle --sourcemap \
		$(OPTIONAL_WEB_BUILD_ARGS)

.PHONY: web
web: $(STATICS) ## Build web-files. (bundle, minify, transpile, etc.)

build/$(APP_NAME): $(STATICS) $(HTML_SOURCES)

## resolve depandancy
OPTIONAL_CLEAN += node_modules

build/$(APP_NAME): build/yarn-updated
build/yarn-updated: package.json
	@$(MAKE) build/tools/yarn
	@mkdir -p $(dir $@)
	yarn install
	touch $@

build/docker-image: package.json

build-tools: build/tools/npm build/tools/yarn build/tools/npx
build/tools/npm:
	@which $(notdir $@)
build/tools/npx:
	@which $(notdir $@)
build/tools/yarn: build/tools/npm
	@which $(notdir $@) || (npm install -g yarn)
