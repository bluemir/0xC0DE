## FE sources
JS_SOURCES    := $(shell find web/js             -type f -name '*.js'   -print)
CSS_SOURCES   := $(shell find web/css            -type f -name '*.css'  -print)
WEB_LIBS      := $(shell find web/lib            -type f                -print)
HTML_SOURCES  := $(shell find web/html-templates -type f -name '*.html' -print)

.watched_sources: $(JS_SOURCES) $(CSS_SOURCES) $(WEB_LIBS) $(HTML_SOURCES)

STATICS :=
STATICS += $(JS_SOURCES:web/%=build/static/%)
STATICS += $(CSS_SOURCES:web/%=build/static/%)
STATICS += $(WEB_LIBS:web/%=build/static/%)

## Static files
build/static/%: web/%
	@mkdir -p $(dir $@)
	cp $< $@

build/$(APP_NAME): $(HTML_SOURCES) $(STATICS)

OPTIONAL_CLEAN_DIR += node_modules

build/$(APP_NAME): build/yarn-updated
build/yarn-updated: package.json yarn.lock
	@$(MAKE) build/tools/yarn
	yarn install
	touch $@

.watched_sources: package.json yarn.lock

build-tools: build/tools/npm build/tools/yarn build/tools/npx
build/tools/npm:
	@which $(notdir $@)
build/tools/npx:
	@which $(notdir $@)
build/tools/yarn: build/tools/npm
	@which $(notdir $@) || (npm install -g yarn)


##### other tools

## roll up
#STATICS := $(filter-out build/static/js/%.js,$(STATICS)) # remove not entrypoint
#STATICS += build/static/js/index.js                      # entrypoint
#build/static/js/%: $(JS_SOURCES) build/yarn-updated
#	@$(MAKE) build/tools/rollup build/tools/npx
#	@mkdir -p $(dir $@)
#	npx rollup $(@:build/%=%) --file $@ --format es -m -p '@rollup/plugin-node-resolve'


## less
#LESS_SOURCES  = $(shell find web/less           -type f -name '*.less' -print)
#STATICS := $(filter-out build/static/css/%,$(STATICS)) # remove default css files
#STATICS += $(LESS_SOURCES:web/less/%=build/static/css/%)
#build/static/css/%: web/less/% build/yarn-updated
#	@$(MAKE) build/tools/lessc
#	@mkdir -p $(dir $@)
#	npx lessc $< $@
#.watched_sources: $(LESS_SOURCES)


#build-tools: build/tools/rollup
build/tools/rollup: build/tools/npm
	@which $(notdir $@) || (npm install -g rollup && npm install -g '@rollup/plugin-node-resolve')

#build-tools: build/tools/lessc
build/tools/lessc: build/tools/npm
	@which $(notdir $@) || (npm install -g less)

