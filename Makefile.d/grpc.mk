PROTO_SOURCE = $(shell find api -type d -name google -prune -o \
                                -type f -name '*.proto' -print)
GO_SOURCES += $(PROTO_SOURCE:%.proto=pkg/gen/%.pb.go)
GO_SOURCES += $(PROTO_SOURCE:%.proto=pkg/gen/%_grpc.pb.go)

OPTIONAL_CLEAN_DIR += pkg/gen

pkg/gen/%.pb.go pkg/gen/%_grpc.pb.go: %.proto $(PROTO_SOURCE)
	@$(MAKE) \
		build/tools/protoc \
		build/tools/protoc-gen-go \
		build/tools/protoc-gen-go-grpc
	@mkdir -p pkg/gen
	protoc \
		-I . \
		--go_out      pkg/gen \
		--go_opt      paths=source_relative \
		--go-grpc_out pkg/gen \
		--go-grpc_opt paths=source_relative \
		$<
