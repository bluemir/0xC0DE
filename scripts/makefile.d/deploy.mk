##@ Deployments
deploy: build/docker-image.pushed ## Deploy webapp
	#@$(MAKE) build/tools/kubectl
	# deploy code
	# example:
	#   cat deployment/server.yaml | DEPLOY_IMAGE=$(shell cat $<) envsubst | kubectl apply -f -
	#   kubectl kustermize deploy | kubectl apply -f -

.PHONY: deploy
tools: build/tools/kubectl
build/tools/kubectl:
	@which $(notdir $@) || (./scripts/tools/install/kubectl.sh)
	#install kubectl. https://kubernetes.io/docs/tasks/tools/

