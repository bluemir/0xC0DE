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

cert-secrets: ## make k8s secret file
cert-secrets: runtime/deploy/local/server-certs.yaml

runtime/deploy/local/server-certs.yaml: CA_CERT=$(CERT_DIR)/local/ca.crt
runtime/deploy/local/server-certs.yaml: CERT=$(CERT_DIR)/local/app/server

runtime/deploy/%-certs.yaml:
	@mkdir -p $(@D)
	@if [ "$(CA_CERT)" == "" ] ; then echo "CA_CERT must be provideded.";  exit 1 ; fi
	@if [ "$(CERT)" == "" ]   ; then echo "CERT must be provideded.";  exit 1 ; fi
	$(MAKE) $(CERT).crt $(CA_CERT)
	kubectl create secret generic $(APP_NAME)-$*-cert \
		--from-file=tls.crt=$(CERT).crt \
		--from-file=tls.key=$(CERT).key \
		--from-file=ca.crt=$(CA_CERT) \
		--dry-run -o yaml \
		> $@

.PHONY: deploy
tools: build/tools/kubectl
build/tools/kubectl:
	@which $(@F) || (./scripts/tools/install/kubectl.sh)
	#install kubectl. https://kubernetes.io/docs/tasks/tools/
	touch $@

