
## go
tools: build/tools/go build/tools/rice
build/tools/go:
	@which $(notdir $@)
build/tools/rice: build/tools/go
	@which $(notdir $@) || (go get -u github.com/GeertJohan/go.rice/rice)

## node.js
tools: build/tools/npm build/tools/yarn
#tools: build/tools/lessc build/tools/rollup
build/tools/npm:
	@which $(notdir $@)
build/tools/rollup: build/tools/npm
	@which $(notdir $@) || (npm install -g rollup && npm install -g '@rollup/plugin-node-resolve')
build/tools/lessc: build/tools/npm
	@which $(notdir $@) || (npm install -g less)
build/tools/yarn: build/tools/npm
	@which $(notdir $@) || (npm install -g yarn)

## grpc
#tools: build/tools/protoc build/tools/protoc-gen-go build/tools/protoc-gen-go-grpc
build/tools/protoc:
	@which $(notdir $@)
build/tools/protoc-gen-go: build/tools/go
	@which $(notdir $@) || (go get -u google.golang.org/protobuf/cmd/protoc-gen-go)
build/tools/protoc-gen-go-grpc: build/tools/go
	@which $(notdir $@) || (go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc)


.PHONY: tools
