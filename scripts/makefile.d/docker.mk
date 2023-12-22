##@ Docker
DOCKER_IMAGE_NAME=$(shell echo $(APP_NAME)| tr A-Z a-z)

docker: build/docker-image ## Build docker image

build/docker-image: Dockerfile $(MAKEFILE_LIST)
	@$(MAKE) build/tools/docker
	@mkdir -p $(dir $@)
	docker build \
		--build-arg VERSION=$(VERSION) \
		-t $(DOCKER_IMAGE_NAME):$(VERSION) \
		-f $< .
	echo $(DOCKER_IMAGE_NAME):$(VERSION) > $@

docker-push: build/docker-image.pushed ## Push docker image

build/docker-image.pushed: build/docker-image
	@$(MAKE) build/tools/docker
	@mkdir -p $(dir $@)
	docker push $(shell cat $<)
	echo $(shell cat $<) > $@

docker-run: build/docker-image ## Run docker container
	docker run -it --rm -v $(PWD)/runtime:/var/run/config $(shell cat $<) $(APP_NAME) -vvvv server

tools: build/tools/docker
build/tools/docker:
	@which $(notdir $@) || (echo "see https://docs.docker.com/engine/install/")

tools: tools/bin/img

tools/bin/img: OS?=$(shell go env GOOS)
tools/bin/img: ARCH?=$(shell go env GOARCH)
tools/bin/img: VERSION=v0.5.11
tools/bin/img:
	@which $(notdir $@) || (mkdir -p $(dir $@) && curl -fSL "https://github.com/genuinetools/img/releases/download/$(VERSION)/img-$(OS)-$(ARCH)" -o "$@" && chmod +x "$@")

.PHONY: docker docker-push docker-run
