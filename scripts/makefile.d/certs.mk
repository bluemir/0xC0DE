##@ Cert

certs: runtime/certs/server-cert.yaml ## Generate self signed certs

.PRECIOUS: runtime/certs/%.key runtime/certs/%.csr runtime/certs/%.crt

runtime/certs/ca.key:
	@mkdir -p $(dir $@)
	openssl genrsa -out $@ 2048
runtime/certs/ca.crt: runtime/certs/ca.key
	@mkdir -p $(dir $@)
	openssl req -new -x509 -days 3650 -key $< \
		-subj "/C=AU/CN=$(APP_NAME)"\
		-out $@

runtime/certs/%.key:
	@mkdir -p $(dir $@)
	openssl genrsa -out $@ 2048
runtime/certs/%.csr: runtime/certs/%.key
	@mkdir -p $(dir $@)
	openssl req -new -key $< \
		-subj "/C=AU/CN=$(APP_NAME)" \
		-out $@
runtime/certs/%.crt: runtime/certs/%.csr runtime/certs/ca.crt runtime/certs/ca.key
	@mkdir -p $(dir $@)
	openssl x509 -req \
		-days 3650 \
		-in $< \
		-CA runtime/certs/ca.crt \
		-CAkey runtime/certs/ca.key \
		-CAcreateserial \
		-out $@ \
		-extfile <(printf "subjectAltName=DNS:$(APP_NAME),DNS:localhost")


runtime/certs/ca.yaml: runtime/certs/ca.crt
	@mkdir -p $(dir $@)
	kubectl create secret generic $(APP_NAME)-ca \
		--from-file=$< \
		--dry-run -o yaml \
		> $@

runtime/certs/%-cert.yaml: runtime/certs/%.crt runtime/certs/%.key runtime/certs/ca.crt
	@mkdir -p $(dir $@)
	kubectl create secret generic $(APP_NAME)-$*-cert \
		--from-file=tls.crt=runtime/certs/$*.crt \
		--from-file=tls.key=runtime/certs/$*.key \
		--from-file=ca.crt=runtime/certs/ca.crt \
		--dry-run -o yaml \
		> $@

