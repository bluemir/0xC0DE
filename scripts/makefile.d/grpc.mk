##@ gRPC
PROTO_SOURCE = $(shell find api/proto -type d -name google -prune -o \
                                      -type f -name '*.proto' -print)

OPTIONAL_CLEAN += $(shell find pkg/api -type f -name '*.go' ! -name 'docs.go')

#pkg/gen/%.pb.go pkg/gen/%_grpc.pb.go pkg/gen/%.pb.gw.go: proto/%.proto $(PROTO_SOURCE)

build/$(APP_NAME).unpacked: build/proto_generated
vet: build/proto_generated
test: build/proto_generated


export PATH:=./runtime/tools/protobuf/bin:$(PATH)

.PHONY: grpc-gen
grpc-gen: build/proto_generated ## Generate grpc codes

build/proto_generated: $(PROTO_SOURCE)
build/proto_generated: | runtime/tools/protoc
build/proto_generated: | runtime/tools/protoc-gen-go runtime/tools/protoc-gen-go-grpc
build/proto_generated: | runtime/tools/protoc-gen-grpc-gateway runtime/tools/protoc-gen-openapiv2
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

## grpc (protoc 플러그인은 PATH 필요 → go build -o 로 go.mod 버전 고정)
build-tools: | runtime/tools/protoc runtime/tools/protoc-gen-go runtime/tools/protoc-gen-go-grpc
runtime/tools/protoc: ./scripts/tools/install/protoc.sh
	@which $(@F) || ($<)
runtime/tools/protoc-gen-go: go.mod go.sum
	go build -o $@ google.golang.org/protobuf/cmd/protoc-gen-go
runtime/tools/protoc-gen-go-grpc: go.mod go.sum
	go build -o $@ google.golang.org/grpc/cmd/protoc-gen-go-grpc

build-tools: runtime/tools/protoc-gen-grpc-gateway
runtime/tools/protoc-gen-grpc-gateway: go.mod go.sum
	go build -o $@ github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway

build-tools: runtime/tools/protoc-gen-openapiv2
runtime/tools/protoc-gen-openapiv2: go.mod go.sum
	go build -o $@ github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

