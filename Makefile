VERSION?=$(shell git describe --tags --dirty --always)
export VERSION

IMPORT_PATH=$(shell cat go.mod | head -n 1 | awk '{print $$2}')
APP_NAME=$(notdir $(IMPORT_PATH))

DOCKER_IMAGE_NAME=$(shell echo $(APP_NAME)| tr A-Z a-z)

export GO111MODULE=on
export GIT_TERMINAL_PROMPT=1

## Go Sources
GO_SOURCES = $(shell find . -name "vendor"  -prune -o \
                            -type f -name "*.go" -print)

## FE sources
JS_SOURCES    = $(shell find static/js             -type f -name '*.js'   -print)
CSS_SOURCES   = $(shell find static/css            -type f -name '*.css'  -print)
WEB_LIBS      = $(shell find static/lib            -type f                -print)
HTML_SOURCES  = $(shell find static/html-templates -type f -name '*.html' -print)

STATICS =
STATICS += build/static/js/bundle.js
STATICS += $(CSS_SOURCES:static/css/%=build/static/css/%)
STATICS += $(WEB_LIBS:static/lib/%=build/static/lib/%)

default: build

## Web dist
build/static/%: static/%
	@mkdir -p $(dir $@)
	cp $< $@

build/static/js/bundle.js: $(JS_SOURCES) build/tools/yarn build/tools/rollup
	@mkdir -p $(dir $@)
	yarn install
	rollup static/js/index.js --file $@ --format es -p '@rollup/plugin-node-resolve'

build: build/$(APP_NAME)

build/$(APP_NAME).unpacked: $(GO_SOURCES) Makefile
	@mkdir -p build
	go build -v \
		-trimpath \
		-ldflags "\
			-X main.AppName=$(APP_NAME) \
			-X main.Version=$(VERSION)  \
		" \
		$(OPTIONAL_BUILD_ARGS) \
		-o $@ main.go
build/$(APP_NAME): build/$(APP_NAME).unpacked $(HTML_SOURCES) $(STATICS) build/tools/rice
	@mkdir -p build
	cp $< $@.tmp
	rice append -v \
		-i $(IMPORT_PATH)/pkg/static \
		--exec $@.tmp
	mv build/$(APP_NAME).tmp $@

docker: build/docker-image

build/docker-image: build/Dockerfile $(GO_SOURCES) $(HTML_SOURCES) $(JS_SOURCES) $(CSS_SOURCES) $(WEB_LIBS)
	docker build \
		--build-arg VERSION=$(VERSION) \
		-t $(DOCKER_IMAGE_NAME):$(VERSION) \
		-f $< .
	echo $(DOCKER_IMAGE_NAME):$(VERSION) > $@

build/Dockerfile: export APP_NAME:=$(APP_NAME)
build/Dockerfile: Dockerfile.template
	@mkdir -p build
	cat $< | envsubst '$${APP_NAME}' > $@

push: build/docker-image.pushed

build/docker-image.pushed: build/docker-image
	docker push $(shell cat $<)
	echo $(shell cat $<) > $@

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

.sources:
	@echo \
	Makefile \
	$(GO_SOURCES) \
	$(JS_SOURCES) \
	$(CSS_SOURCES) \
	$(WEB_LIBS) \
	$(HTML_SOURCES) \
	| tr " " "\n"

test:
	go test -v ./pkg/...

reset:
	ps -e | grep make | grep -v grep | awk '{print $$1}' | xargs kill

deploy: build/docker-image.pushed

# TOOLS
-include makefile.d/*

.PHONY: build docker push clean run auto-run .sources test deploy reset
