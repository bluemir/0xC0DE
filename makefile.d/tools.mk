tools: build/tools/go build/tools/rice build/tools/npm build/tools/rollup build/tools/yarn

build/tools/go:
	which go
build/tools/rice: build/tools/go
	which rice || (GO111MODULE=off go get -u github.com/GeertJohan/go.rice/rice)
build/tools/npm:
	which npm
build/tools/rollup: build/tools/npm
	which rollup || npm install -g rollup && npm install -g '@rollup/plugin-node-resolve'
build/tools/yarn: build/tools/npm
	which yarn || npm install -g yarn

.PHONY: tools
