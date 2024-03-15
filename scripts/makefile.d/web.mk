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
STATICS += $(CSS_SOURCES:assets/%=build/static/%)
STATICS += $(WEB_LIBS:assets/%=build/static/%)
STATICS += $(IMAGES:assets/%=build/static/%)
STATICS += $(WEB_META:assets/%=build/static/%)

build/static/%: assets/%
	@mkdir -p $(dir $@)
	cp $< $@

## js import helper
build/static/js/index.js: assets/js/elements/index.js assets/js/layout/index.js assets/js/components/index.js
assets/js/%/index.js: $(shell find assets/js/ -type f -name '*.js' -print) scripts/tools/import-helper.sh
	 scripts/tools/import-helper.sh $(dir $@) > $@
OPTIONAL_CLEAN += assets/js/elements/index.js assets/js/layout/index.js assets/js/components/index.js

## esbuild
STATICS += build/static/js/index.js # entrypoint
build/static/js/%: export NODE_PATH=assets/js:assets/lib
build/static/js/%: $(JS_SOURCES) build/yarn-updated
	@$(MAKE) build/tools/npx
	@mkdir -p $(dir $@)
	npx esbuild $(@:build/static/%=assets/%) --outdir=$(dir $@) \
		--bundle --sourcemap --format=esm \
		--external:bm.js --external:lit-html \
		$(OPTIONAL_WEB_BUILD_ARGS)
	#--external:/config.js \
	#--minify \

## rollup & js
## yarn add --dev rollup '@rollup/plugin-node-resolve'
#STATICS := $(filter-out build/static/js/%.js,$(STATICS)) # remove not entrypoint
#STATICS += build/static/js/index.js                      # entrypoint
#build/static/js/%: $(JS_SOURCES) build/yarn-updated
#	@$(MAKE) build/tools/npx
#	@mkdir -p $(dir $@)
#	npx rollup $(@:build/static/%=assets/%) --file $@ --format es -m -p '@rollup/plugin-node-resolve'

## less
## yarn add --dev less
#LESS_SOURCES  = $(shell find assets/less           -type f -name '*.less' -print)
#STATICS := $(filter-out build/static/css/%,$(STATICS)) # remove default css files
#STATICS += $(LESS_SOURCES:assets/less/%=build/static/css/%)
#build/static/css/%: assets/less/% build/yarn-updated
#	@$(MAKE) build/tools/npx
#	@mkdir -p $(dir $@)
#	npx lessc $< $@

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
