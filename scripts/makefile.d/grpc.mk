##@ gRPC
PROTO_SOURCE = $(shell find api/proto -type d -name google -prune -o \
                                      -type f -name '*.proto' -print)

OPTIONAL_CLEAN += $(shell find pkg/api -type f -name '*.go' ! -name 'docs.go')

#pkg/gen/%.pb.go pkg/gen/%_grpc.pb.go pkg/gen/%.pb.gw.go: proto/%.proto $(PROTO_SOURCE)

build/$(APP_NAME).unpacked: build/proto_generated
vet: build/proto_generated
test: build/proto_generated


export PATH:=./build/tools/protobuf/bin:$(PATH)

.PHONY: grpc-gen
grpc-gen: build/proto_generated ## Generate grpc codes

build/proto_generated: $(PROTO_SOURCE) build/tools/protoc 
	@$(MAKE) \
		build/tools/protoc \
		build/tools/protoc-gen-go \
		build/tools/protoc-gen-go-grpc \
		build/tools/protoc-gen-grpc-gateway \
		build/tools/protoc-gen-openapiv2
	@mkdir -p pkg/ build/openapiv2
	protoc \
		-I api/proto \
		--go_out           pkg \
		--go_opt           paths=source_relative \
		--go-grpc_out      pkg \
		--go-grpc_opt      paths=source_relative \
		--grpc-gateway_out pkg \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt generate_unbound_methods=true \
		--openapiv2_out    build/openapiv2 \
		$(PROTO_SOURCE)
	touch $@

build/docker-image: $(PROTO_SOURCE)

## grpc
build-tools: | build/tools/protoc build/tools/protoc-gen-go build/tools/protoc-gen-go-grpc
build/tools/protoc: ./scripts/tools/install/protoc.sh
	@which $(@F) || ($<)
build/tools/protoc-gen-go: ./scripts/tools/install/go-tool.sh |  build/tools/go
	@which $(@F) || ($< google.golang.org/protobuf/cmd/protoc-gen-go)
build/tools/protoc-gen-go-grpc: ./scripts/tools/install/go-tool.sh | build/tools/go
	@which $(@F) || ($< google.golang.org/grpc/cmd/protoc-gen-go-grpc)

build-tools: build/tools/protoc-gen-grpc-gateway
build/tools/protoc-gen-grpc-gateway: | build/tools/go
	@which $(@F) || (./scripts/tools/install/go-tool.sh github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway)

build-tools: build/tools/protoc-gen-openapiv2
build/tools/protoc-gen-openapiv2: | build/tools/go
	@which $(@F) || (./scripts/tools/install/go-tool.sh github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2)

