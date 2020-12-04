tools: build/tools/go build/tools/rice build/tools/npm build/tools/yarn
#tools: build/tools/lessc build/tools/rollup

build/tools/go:
	@which $(notdir $@)
build/tools/rice: build/tools/go
	@which $(notdir $@) || (go install github.com/GeertJohan/go.rice/rice)
build/tools/npm:
	@which $(notdir $@)
build/tools/rollup: build/tools/npm
	@which $(notdir $@) || (npm install -g rollup && npm install -g '@rollup/plugin-node-resolve')
build/tools/lessc: build/tools/npm
	@which $(notdir $@) || (npm install -g less)
build/tools/yarn: build/tools/npm
	@which $(notdir $@) || (npm install -g yarn)

.PHONY: tools
