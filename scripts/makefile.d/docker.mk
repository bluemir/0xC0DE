##@ Docker
IMAGE_NAME=$(shell echo $(APP_NAME)| tr A-Z a-z)

docker: build/docker-image ## Build docker image

#build/docker-image: export DOCKER_BUILDKIT=1
build/docker-image: Dockerfile $(MAKEFILE_LIST)
	@$(MAKE) build/tools/docker
	@mkdir -p $(dir $@)
	docker build \
		--build-arg VERSION=$(VERSION) \
		-t $(MAGE_NAME):$(VERSION) \
		-f $< .
	echo $(IMAGE_NAME):$(VERSION) > $@

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

.PHONY: docker docker-push docker-run
