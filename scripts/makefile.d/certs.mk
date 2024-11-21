##@ Cert

.PHONY: certs
certs: ## Generate self signed certs
#certs: $(CERT_DIR)/local/app/server.crt $(CERT_DIR)/local/app/server.bundle.crt
#certs: $(CERT_DIR)/local/app/server.crt $(CERT_DIR)/local/app/server.bundle.crt
#certs: $(CERT_DIR)/local/etcd/server.crt
#certs: $(CERT_DIR)/local/buildkitd/server.crt


# customize SAN via OPTIONAL_SAN
# eg.
# $(CERT_DIR)/local/server.crt: export OPTIONAL_SAN=",DNS:dev.0xC0DE.io"

cert-secrets: ## make k8s secret file
#cert-secrets: runtime/deploy/server.yaml
#cert-secrets: runtime/deploy/server.bundle.yaml

##########################################################################


.PRECIOUS: $(CERT_DIR)/%/ca.key $(CERT_DIR)/%/ca.crt
$(CERT_DIR)/%/ca.key:
	@mkdir -p $(@D)
	openssl genrsa -out $@ 2048
$(CERT_DIR)/%/ca.crt: $(CERT_DIR)/%/ca.key
	@mkdir -p $(@D)
	openssl req -new -x509 -days 3650 -key $< \
		-subj "/C=AU/CN=$(APP_NAME)"\
		-out $@

.PRECIOUS: $(CERT_DIR)/%.key $(CERT_DIR)/%.csr $(CERT_DIR)/%.crt
.PRECIOUS: $(CERT_DIR)/%.bundle.crt
$(CERT_DIR)/%.key: $(MAKEFILES)
	@mkdir -p $(@D)
	openssl genrsa -out $@ 2048
$(CERT_DIR)/%.csr: $(CERT_DIR)/%.key $(MAKEFILES)
	@mkdir -p $(@D)
	openssl req -new -key $< \
		-subj "/C=AU/CN=$(APP_NAME)" \
		-out $@
$(CERT_DIR)/%.crt: $(CERT_DIR)/%.csr
	@mkdir -p $(@D)
	# ca cert: $(filter %.crt,$^)
	@if [ "$(filter %.crt,$^)" == "" ] ;          then echo "ca cert not found";   exit 1; fi
	@if [ "$(words $(filter %.crt,$^))" -ne 1 ] ; then echo "ca cert must be one"; exit 1; fi
	# ca key:  $(patsubst %.crt,%.key,$(filter %.crt,$^))
	openssl x509 -req \
		-days 3650 \
		-in $< \
		-CA $(filter %.crt,$^) \
		-CAkey $(patsubst %.crt,%.key,$(filter %.crt,$^)) \
		-CAcreateserial \
		-out $@ \
		-extfile <(printf "subjectAltName=DNS:$(APP_NAME),DNS:localhost$(OPTIONAL_SAN)")
	################################# cert issued #################################
	# name:         $@
	# check cert:   openssl x509 -text -noout -in $@
	# check issuer: openssl x509 -subject -issuer -noout -in $@
	# check SAN:    openssl x509 -text -noout -in $@ | grep "Subject Alternative Name" -A1
	################################################################################

$(CERT_DIR)/%.bundle.crt:
	cat $^ > $@

$(CERT_DIR)/%.yaml: build/tools/kubectl
	@mkdir -p $(@D)
	@if [ "$(CERT)" == "" ] ;     then echo "CERT MUST declare";   exit 1; fi
	@if [ "$(CERT_KEY)" == "" ] ; then echo "CERT_KEY MUST declare";   exit 1; fi
	@if [ "$(CA_CERT)" == "" ] ;  then echo "CA_CERT MUST declare";   exit 1; fi
	kubectl create secret generic $(subst /,-,$*) \
		--from-file=tls.crt=$(CERT) \
		--from-file=tls.key=$(CERT_KEY) \
		--from-file=ca.crt=$(CA_CERT) \
		--dry-run -o yaml \
		> $@
