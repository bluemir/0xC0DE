PROTO_SOURCE = $(shell find proto -type d -name google -prune -o \
                                  -type f -name '*.proto' -print)

OPTIONAL_CLEAN_DIR += pkg/gen

#pkg/gen/%.pb.go pkg/gen/%_grpc.pb.go pkg/gen/%.pb.gw.go: proto/%.proto $(PROTO_SOURCE)

build/$(APP_NAME).unpacked: build/proto_generated
test: build/proto_generated

build/proto_generated: $(PROTO_SOURCE)
	@$(MAKE) \
		build/tools/protoc \
		build/tools/protoc-gen-go \
		build/tools/protoc-gen-go-grpc \
		build/tools/protoc-gen-grpc-gateway \
		build/tools/protoc-gen-openapiv2
	@mkdir -p pkg/gen build/openapiv2
	protoc \
		-I proto \
		--go_out           pkg/gen \
		--go_opt           paths=source_relative \
		--go-grpc_out      pkg/gen \
		--go-grpc_opt      paths=source_relative \
		--grpc-gateway_out pkg/gen \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt generate_unbound_methods=true \
		--openapiv2_out    build/openapiv2 \
		$(PROTO_SOURCE)
	touch $@

.watched_sources: $(PROTO_SOURCE)
