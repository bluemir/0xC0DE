##@ etcd

ETCD_DATA_DIR=runtime/etcd.data

run-etcd: ## Run etcd for development
run-etcd: runtime/certs/local/etcd/server.crt runtime/certs/local/etcd/server.key
run-etcd: runtime/certs/local/etcd-client/ca.crt
run-etcd: | runtime/tools/etcd
	etcd \
		--name local \
		--cert-file       runtime/certs/local/etcd/server.crt \
		--key-file        runtime/certs/local/etcd/server.key \
		--trusted-ca-file runtime/certs/local/etcd/ca.crt \
		--client-cert-auth \
		--advertise-client-urls 'https://127.0.0.1:2379' \
		--listen-client-urls    'https://0.0.0.0:2379' \
		--data-dir $(ETCD_DATA_DIR)

# trusted-ca-file은 서버 스스로의 cert를 검증하는데도 쓰고 client cert 를 검증하는데도 쓴다..?

run-etcd-client: | runtime/tools/etcdctl
run-etcd-client: runtime/certs/local/etcd/client.crt runtime/certs/local/etcd/client.key
run-etcd-client: runtime/certs/local/etcd/ca.crt
	etcdctl \
		--cert   runtime/certs/local/etcd/client.crt \
		--key    runtime/certs/local/etcd/client.key \
		--cacert runtime/certs/local/etcd/ca.crt \
		--endpoints https://127.0.0.1:2379 \
		put "ping" "pong"

build-tools: runtime/tools/etcd
runtime/tools/etcd runtime/tools/etcdctl: runtime/tools/go
	@which $(notdir $@) || (./scripts/tools/install/etcd.sh v3.5.11)
