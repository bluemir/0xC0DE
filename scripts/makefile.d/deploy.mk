##@ Deployments
deploy: build/docker-image.pushed ## Deploy webapp
	#@$(MAKE) build/tools/kubectl
	#@if [ "$(IMAGE_PULL_SECRET_NAME)" == "" ] ; then echo "IMAGE_PULL_SECRET_NAME must provideded.";  exit 1 ; fi
	#@if [ "$(NAMESPACE)" == "" ] ; then echo "NAMESPACE must provideded.";  exit 1 ; fi
	#@kubectl get -n $(NAMESPACE) secrets/$(IMAGE_PULL_SECRET_NAME) || (echo 'secrets "$(IMAGE_PULL_SECRET_NAME)" not found.' ; exit 1)
	# deploy code
	# example:
	#   cat deployment/server.yaml \
	#     | DEPLOY_IMAGE=$(shell cat $<) envsubst \
	#     | kubectl apply -f -
	#   kubectl kustermize deploy | kubectl apply -f -

.PHONY: deploy
tools: build/tools/kubectl
build/tools/kubectl:
	@which $(notdir $@) || (./scripts/tools/install/kubectl.sh)
	#install kubectl. https://kubernetes.io/docs/tasks/tools/

