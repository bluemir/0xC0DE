VERSION?=$(shell git describe --tags --dirty --always)
export VERSION

IMPORT_PATH=$(shell cat go.mod | head -n 1 | awk '{print $$2}')
APP_NAME=$(notdir $(IMPORT_PATH))

export GO111MODULE=on
export GIT_TERMINAL_PROMPT=1

## Go Sources
GO_SOURCES = $(shell find . -name "vendor"  -prune -o \
                            -type f -name "*.go" -print)

## FE sources
JS_SOURCES    = $(shell find static/js             -type f -name '*.js'   -print)
CSS_SOURCES   = $(shell find static/css            -type f -name '*.css'  -print)
#LESS_SOURCES  = $(shell find static/less           -type f -name '*.less' -print)
WEB_LIBS      = $(shell find static/lib            -type f                -print)
HTML_SOURCES  = $(shell find static/html-templates -type f -name '*.html' -print)

STATICS =
STATICS += $(JS_SOURCES:%=build/%)
#STATICS += build/static/js/index.js   # for rollup
STATICS += $(CSS_SOURCES:%=build/%)
#STATICS += $(LESS_SOURCES:static/less/%=build/static/css/%)  # for less
STATICS += $(WEB_LIBS:%=build/%)

default: build

## Web dist
build/static/%: static/%
	@mkdir -p $(dir $@)
	cp $< $@

#build/static/js/%: $(JS_SOURCES) build/yarn-updated
#	@$(MAKE) build/tools/rollup
#	@mkdir -p $(dir $@)
#	rollup $(@:build/%=%) --file $@ --format es -p '@rollup/plugin-node-resolve'

#build/static/css/%: static/less/% build/yarn-updated
#	@$(MAKE) build/tools/lessc
#	@mkdir -p $(dir $@)
#	lessc $< $@

build/yarn-updated: yarn.lock package.json
	@$(MAKE) build/tools/yarn
	yarn install
	touch $@

build: build/$(APP_NAME)

build/$(APP_NAME).unpacked: $(GO_SOURCES) Makefile
	@$(MAKE) build/tools/go
	@mkdir -p build
	go build -v \
		-trimpath \
		-ldflags "\
			-X main.AppName=$(APP_NAME) \
			-X main.Version=$(VERSION)  \
		" \
		$(OPTIONAL_BUILD_ARGS) \
		-o $@ main.go
build/$(APP_NAME): build/$(APP_NAME).unpacked $(HTML_SOURCES) $(STATICS) build/yarn-updated
	$(MAKE) build/tools/rice
	@mkdir -p build
	cp $< $@.tmp
	rice append -v \
		-i $(IMPORT_PATH)/pkg/static \
		--exec $@.tmp
	mv build/$(APP_NAME).tmp $@

clean:
	rm -rf build/ node_modules/

run: build/$(APP_NAME)
	$< -vvvv server

auto-run:
	while true; do \
		$(MAKE) .sources | \
		entr -rd $(MAKE) test run ;  \
		echo "hit ^C again to quit" && sleep 1  \
	; done

reset:
	ps -e | grep make | grep -v grep | awk '{print $$1}' | xargs kill

.sources:
	@echo \
	Makefile \
	go.mod go.sum \
	package.json yarn.lock \
	$(GO_SOURCES) \
	$(JS_SOURCES) \
	$(CSS_SOURCES) \
	$(WEB_LIBS) \
	$(HTML_SOURCES) \
	| tr " " "\n"

test:
	go test -v ./pkg/...

# sub-makefiles
# for build tools, docker build, deploy
-include Makefile.d/*

.PHONY: build clean run auto-run reset .sources test
