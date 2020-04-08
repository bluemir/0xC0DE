##########################
# DONOT modify this file #
# using makefile.d/      #
##########################

VERSION?=$(shell git describe --long --tags --dirty --always)
export VERSION

IMPORT_PATH=$(shell cat go.mod | head -n 1 | awk '{print $$2}')
APP_NAME=$(notdir $(IMPORT_PATH))

export GO111MODULE=on
export GIT_TERMINAL_PROMPT=1

DOCKER_IMAGE_NAME=bluemir/$(APP_NAME)

## Go Sources
GO_SOURCES = $(shell find . -name "vendor"  -prune -o \
                            -type f -name "*.go" -print)

## resources
RESOURCES     = $(shell find resources/     -type f                -print)

## FE sources
JS_SOURCES    = $(shell find resources/js   -type f -name '*.js'   -print)
CSS_SOURCES   = $(shell find resources/css  -type f -name '*.css'  -print)
WEB_LIBS      = $(shell find resources/lib  -type f                -print)

STATICS =
STATICS += $(JS_SOURCES:resources/js/%=build/static/js/%)
STATICS += $(CSS_SOURCES:resources/css/%=build/static/css/%)
STATICS += $(WEB_LIBS:resources/lib/%=build/static/lib/%)

default: build

-include makefile.d/*

## Web dist
build/static/%: resources/%
	@mkdir -p $(dir $@)
	cp $< $@

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
build/$(APP_NAME): build/$(APP_NAME).unpacked $(RESOURCES) $(STATICS)
	@$(MAKE) build/tools/rice
	@mkdir -p build
	cp $< $@.tmp
	rice append -v \
		-i $(IMPORT_PATH)/pkg/resources \
		--exec $@.tmp
	mv build/$(APP_NAME).tmp $@

docker: build/docker-image

build/docker-image: build/Dockerfile $(GO_SOURCES) $(HTML_SOURCES) $(RESOURCES) $(STATICS)
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
	rm -rf build/

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
	$(RESOURCES) \
	| tr " " "\n"

test:
	go test -v ./pkg/...

deploy: build/docker-image.pushed

# TOOLS
tools: build/tools/rice

build/tools/rice:
	which rice || (GO111MODULE=off go get -u github.com/GeertJohan/go.rice/rice)

.PHONY: build docker push clean run auto-run .sources test deploy
