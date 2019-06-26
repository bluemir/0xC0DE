VERSION?=$(shell git describe --long --tags --dirty --always)
export VERSION

IMPORT_PATH:=github.com/bluemir/0xC0DE
BIN_NAME:=$(notdir $(IMPORT_PATH))

export GO111MODULE=on

DOCKER_IMAGE_NAME=bluemir/$(BIN_NAME)

## Go Sources
GO_SOURCES = $(shell find . -name "vendor"  -prune -o \
                            -name "*.pb.go" -prune -o \
                            -type f -name "*.go" -print)

## FE sources
JS_SOURCES    = $(shell find app/js       -type f -name '*.js'   -print)
HTML_SOURCES  = $(shell find app/html     -type f -name '*.html' -print)
CSS_SOURCES   = $(shell find app/css      -type f -name '*.css'  -print)
WEB_LIBS      = $(shell find app/lib      -type f                -print)
HTML_TEMPLATE = $(shell find app/template -type f -name '*.html' -print)

DISTS =
DISTS += $(HTML_SOURCES:app/html/%=build/dist/html/%)
DISTS += $(JS_SOURCES:app/js/%=build/dist/js/%)
DISTS += $(CSS_SOURCES:app/css/%=build/dist/css/%)
DISTS += $(WEB_LIBS:app/lib/%=build/dist/lib/%)

default: build

## Web dist
#dist/css/%.css: $(CSS_SOURCES)
#	lessc app/less/entry/$*.less $@
build/dist/%: app/%
	@mkdir -p $(dir $@)
	cp $< $@

build: build/$(BIN_NAME)

build/$(BIN_NAME).unpacked: $(GO_SOURCES) makefile
	@mkdir -p build
	go build -v \
		-ldflags "-X main.VERSION=$(VERSION)" \
		$(OPTIONAL_BUILD_ARGS) \
		-o $@ main.go
build/$(BIN_NAME): build/$(BIN_NAME).unpacked $(HTML_TEMPLATE) $(DISTS)
	@mkdir -p build
	cp $< $@.tmp
	rice append -v \
		-i $(IMPORT_PATH)/pkg/dist \
		--exec $@.tmp
	mv build/$(BIN_NAME).tmp $@

docker: build/.docker-image

build/.docker-image: $(GO_SOURCES) $(HTML_TEMPLATE) $(DISTS) Dockerfile
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BIN_NAME=$(BIN_NAME) \
		-t $(DOCKER_IMAGE_NAME):$(VERSION) .
	echo $(DOCKER_IMAGE_NAME):$(VERSION) > $@

push: build/.docker-image.pushed

build/.docker-image.pushed: .docker-image
	docker push $(shell cat build/.docker-image)
	echo $(shell cat build/.docker-image) > $@

clean:
	rm -rf .docker-image .docker-image.pushed build/

run: export LOG_LEVEL=TRACE
run: build/$(BIN_NAME)
	$<

auto-run:
	while true; do \
		$(MAKE) .sources | \
		entr -rd $(MAKE) test run ;  \
		echo "hit ^C again to quit" && sleep 1  \
	; done

.sources:
	@echo \
	makefile \
	$(GO_SOURCES) \
	$(JS_SOURCES) \
	$(HTML_SOURCES) \
	$(CSS_SOURCES) \
	$(WEB_LIBS) \
	$(HTML_TEMPLATE) \
	| tr " " "\n"

test:
	go test -v ./pkg/...

.PHONY: build docker push clean run auto-run .sources test
