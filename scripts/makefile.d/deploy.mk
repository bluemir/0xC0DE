##@ Deployments
#deploy: | runtime/tools/kubectl
deploy: build/docker-image.pushed ## Deploy webapp
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

runtime/deploy/%-certs.yaml: | runtime/tools/kubectl
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
tools: runtime/tools/kubectl
runtime/tools/kubectl:
	@which $(@F) || (./scripts/tools/install/kubectl.sh v1.33.1)
	#install kubectl. https://kubernetes.io/docs/tasks/tools/
	touch $@

runtime/tools/helm:
	@which $(@F) || (./scripts/tools/install/helm.sh v3.18.6)


runtime/tools/yq:
	@mkdir -p $(@D)
	@which $(@F) || (./scripts/tools/install/yq.sh v4.47.1)
