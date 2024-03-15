##@ Cert

CERT_DIR=runtime/certs

certs: ## Generate self signed certs
certs: $(CERT_DIR)/local/app/server.crt $(CERT_DIR)/local/app/server.bundle.crt
certs: $(CERT_DIR)/local/etcd/server.crt

# customize SAN via OPTIONAL_SAN
# eg.
# $(CERT_DIR)/local/server.crt: export OPTIONAL_SAN=",DNS:dev.0xC0DE.io"
$(CERT_DIR)/local/etcd/server.crt: export OPTIONAL_SAN=",DNS:local.bluemir.me,IP:127.0.0.1"
$(CERT_DIR)/local/etcd/client.crt: export OPTIONAL_SAN=",IP:127.0.0.1"


cert-secrets: ## make k8s secret file
#cert-secrets: runtime/deploy/server.bundle.yaml

##########################################################################


.PRECIOUS: $(CERT_DIR)/%/ca.key $(CERT_DIR)/%/ca.crt
$(CERT_DIR)/%/ca.key:
	@mkdir -p $(dir $@)
	openssl genrsa -out $@ 2048
$(CERT_DIR)/%/ca.crt: $(CERT_DIR)/%/ca.key
	@mkdir -p $(dir $@)
	openssl req -new -x509 -days 3650 -key $< \
		-subj "/C=AU/CN=$(APP_NAME)"\
		-out $@

.SECONDEXPANSION:

.PRECIOUS: $(CERT_DIR)/%.key $(CERT_DIR)/%.csr $(CERT_DIR)/%.crt
.PRECIOUS: $(CERT_DIR)/%.bundle.crt
$(CERT_DIR)/%.key: $(MAKEFILES)
	@mkdir -p $(dir $@)
	openssl genrsa -out $@ 2048
$(CERT_DIR)/%.csr: $(CERT_DIR)/%.key $(MAKEFILES)
	@mkdir -p $(dir $@)
	openssl req -new -key $< \
		-subj "/C=AU/CN=$(APP_NAME)" \
		-out $@
$(CERT_DIR)/%.crt: $(CERT_DIR)/%.csr $(CERT_DIR)/$$(*D)/ca.crt $(CERT_DIR)/$$(*D)/ca.key
	@mkdir -p $(dir $@)
	openssl x509 -req \
		-days 3650 \
		-in $< \
		-CA $(@D)/ca.crt \
		-CAkey $(@D)/ca.key \
		-CAcreateserial \
		-out $@ \
		-extfile <(printf "subjectAltName=DNS:$(APP_NAME),DNS:localhost$(OPTIONAL_SAN)")
	################################# cert issued #################################
	# name:         $@
	# check cert:   openssl x509 -text -noout -in $@
	# check issuer: openssl x509 -subject -issuer -noout -in $@
	# check SAN:    openssl x509 -text -noout -in $@ | grep "Subject Alternative Name" -A1
	################################################################################
$(CERT_DIR)/%.bundle.crt: $(CERT_DIR)/%.crt $(CERT_DIR)/$$(*D)/ca.crt
	cat $< $(@D)/ca.crt > $@

runtime/deploy/ca.yaml: $(CERT_DIR)/ca.crt
	@mkdir -p $(dir $@)
	kubectl create secret generic $(APP_NAME)-ca \
		--from-file=$< \
		--dry-run -o yaml \
		> $@

runtime/deploy/%-cert.yaml: $(CERT_DIR)/%.crt $(CERT_DIR)/%.key $(CERT_DIR)/$$(*D)/ca.crt
	@mkdir -p $(dir $@)
	kubectl create secret generic $(APP_NAME)-$*-cert \
		--from-file=tls.crt=$(CERT_DIR)/$*.crt \
		--from-file=tls.key=$(CERT_DIR)/$*.key \
		--from-file=ca.crt=$(CERT_DIR)/ca.crt \
		--dry-run -o yaml \
		> $@

