
OPTIONAL_CLEAN_DIR += node_modules

build/$(APP_NAME): build/yarn-updated
build/yarn-updated: package.json yarn.lock
	@$(MAKE) build/tools/yarn
	yarn install
	touch $@

.sources: package.json yarn.lock


## roll up
#STATICS := $(filter-out build/static/js/%.js,$(STATICS)) # remove not entrypoint
#STATICS += build/static/js/index.js
#build/static/js/%: $(JS_SOURCES) build/yarn-updated
#	@$(MAKE) build/tools/rollup
#	@mkdir -p $(dir $@)
#	rollup $(@:build/%=%) --file $@ --format es -p '@rollup/plugin-node-resolve'

## less
#LESS_SOURCES  = $(shell find static/less           -type f -name '*.less' -print)
#STATICS := $(filter-out build/static/css/%,$(STATICS)) # remove default css files
#STATICS += $(LESS_SOURCES:static/less/%=build/static/css/%)
#build/static/css/%: static/less/% build/yarn-updated
#	@$(MAKE) build/tools/lessc
#	@mkdir -p $(dir $@)
#	lessc $< $@
#.sources: $(LESS_SOURCES)

##### tools
tools: build/tools/npm build/tools/yarn
#tools: build/tools/lessc build/tools/rollup
build/tools/npm:
	@which $(notdir $@)
build/tools/rollup: build/tools/npm
	@which $(notdir $@) || (npm install -g rollup && npm install -g '@rollup/plugin-node-resolve')
build/tools/lessc: build/tools/npm
	@which $(notdir $@) || (npm install -g less)
build/tools/yarn: build/tools/npm
	@which $(notdir $@) || (npm install -g yarn)

