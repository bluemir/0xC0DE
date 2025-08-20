##@ Cert

CERT_DIR=runtime/certs

certs: ## Generate self signed certs
certs: $(CERT_DIR)/local/app/server.key $(CERT_DIR)/local/app/server.bundle.crt
certs: $(CERT_DIR)/local/etcd/server.crt

# customize key size via KEY_SIZE
# eg.
# $(CERT_DIR)/local/server.key: KEY_SIZE=4096
#
# customize SAN via OPTIONAL_SAN
# eg.
# $(CERT_DIR)/local/server.csr: OPTIONAL_SAN=",DNS:example.com,DNS:example.localhost,IP:127.0.0.1"
#
# customize signing chain via SINGING_CERT
# eg.
# $(CERT_DIR)/local/app/server.crt: SINGING_CERT=$(CERT_DIR)/local/ca
#
# customize bundle
# eg.
# $(CERT_DIR)/local/app/server.bundle.crt: $(CERT_DIR)/local/app/server.crt
# $(CERT_DIR)/local/app/server.bundle.crt: $(CERT_DIR)/local/ca.crt

$(CERT_DIR)/local/app/server.csr:  OPTIONAL_SAN=",IP:127.0.0.1"

$(CERT_DIR)/local/etcd/server.csr: OPTIONAL_SAN=",IP:127.0.0.1"
$(CERT_DIR)/local/etcd/server.crt: SINGING_CERT=$(CERT_DIR)/ca

cert-secrets: ## make k8s secret file
cert-secrets: runtime/deploy/local/server-certs.yaml

runtime/deploy/local/server-certs.yaml: CA_CERT=$(CERT_DIR)/local/ca.crt
runtime/deploy/local/server-certs.yaml: CERT=$(CERT_DIR)/local/app/server

##########################################################################

.PRECIOUS: $(CERT_DIR)/%.key $(CERT_DIR)/%.csr $(CERT_DIR)/%.crt
.PRECIOUS: $(CERT_DIR)/%.bundle.crt

$(CERT_DIR)/%.key: KEY_SIZE?=2048
$(CERT_DIR)/%.key: $(MAKEFILES)
	@mkdir -p $(dir $@)
	openssl genrsa -out $@ $(KEY_SIZE)

$(CERT_DIR)/ca.crt: $(CERT_DIR)/ca.key
	@mkdir -p $(dir $@)
	openssl req -new -x509 -days 3650 -key $< \
		-subj "/C=AU/CN=$(APP_NAME)"\
		-addext "basicConstraints=critical,CA:TRUE" \
		-out $@
$(CERT_DIR)/%/ca.crt: $(CERT_DIR)/%/ca.key
	@mkdir -p $(dir $@)
	openssl req -new -x509 -days 3650 -key $< \
		-subj "/C=AU/CN=$(APP_NAME)"\
		-addext "basicConstraints=critical,CA:TRUE" \
		-out $@

$(CERT_DIR)/%.csr: COMMON_NAME?=$(APP_NAME)
$(CERT_DIR)/%.csr: $(CERT_DIR)/%.key $(MAKEFILES)
	@mkdir -p $(dir $@)
	openssl req -new -key $< \
		-subj "/C=AU/CN=$(COMMON_NAME)" \
		$(OPTIONAL_CSR_ARGS) \
		-addext "subjectAltName=DNS:$(COMMON_NAME).localhost,DNS:localhost$(OPTIONAL_SAN)" \
		-out $@

#$(CERT_DIR)/%.crt: SINGING_CERT?=$(CERT_DIR)/ca
$(CERT_DIR)/%.crt: $(CERT_DIR)/%.csr $(MAKEFILES)
	@mkdir -p $(dir $@)
	@if [ "$(SINGING_CERT)" == "" ] ; then echo "SINGING_CERT must be provideded.";  exit 1 ; fi
	$(MAKE) $(SINGING_CERT).crt # ensure signing cert for $@
	openssl x509 -req \
		-days 3650 \
		-in $< \
		-copy_extensions copyall \
		-CA    $(SINGING_CERT).crt \
		-CAkey $(SINGING_CERT).key \
		-CAcreateserial \
		-out $@
	################################# cert issued #################################
	# name:         $@
	# check cert:   openssl x509 -text -noout -in $@
	# check issuer: openssl x509 -subject -issuer -noout -in $@
	# check SAN:    openssl x509 -text -noout -in $@ | grep "Subject Alternative Name" -A1
	################################################################################
$(CERT_DIR)/%.bundle.crt:
	cat $^ > $@
