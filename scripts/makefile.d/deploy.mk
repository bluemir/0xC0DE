deploy: build/docker-image.pushed
	# deploy code
	# cat deploy.yaml | DEPLOY_IMAGE=$(shell cat $<) envsubst | kubectl apply -f -

#tools: build/tools/kubectl
build/tools/kubectl:
	@which $(notdir $@) || (echo "install kubectl")

.PHONY: deploy
