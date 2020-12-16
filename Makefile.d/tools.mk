# Common tools like go-compiler

## go
tools: build/tools/go build/tools/rice
build/tools/go:
	@which $(notdir $@)
build/tools/rice: build/tools/go
	@which $(notdir $@) || (go get -u github.com/GeertJohan/go.rice/rice)

.PHONY: tools
