DOCKER_IMAGE_NAME=$(shell echo $(APP_NAME)| tr A-Z a-z)

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

docker-push: build/docker-image.pushed

build/docker-image.pushed: build/docker-image
	docker push $(shell cat $<)
	echo $(shell cat $<) > $@

.PHONY: docker docker-push
