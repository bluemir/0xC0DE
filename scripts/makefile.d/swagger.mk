##@ Swagger
#.PRECIOUS: internal/swagger/docs.go

# see https://github.com/swaggo/swag for documents

build/$(APP_NAME).unpacked: internal/swagger/docs.go
test: internal/swagger/docs.go
vet: internal/swagger/docs.go

OPTIONAL_CLEAN+= internal/swagger/docs.go internal/swagger/swagger.json internal/swagger/swagger.yaml

.PHONY: swagger
swagger: internal/swagger/docs.go ## Make swagger file

internal/swagger/docs.go: $(filter ./internal/server/%.go,$(GO_SOURCES))
	@$(MAKE) build/tools/swag
	@mkdir -p $(dir $@)
	swag init \
		--generalInfo internal/server/routes.go \
		--parseInternal \
		--output $(dir $@)
	# for dependency add this option: `--parseDependency`

build/tools/swag:
	@$(MAKE) build/tools/go
	@which $(notdir $@) || (./scripts/makefile.d/install-go-tool.sh github.com/swaggo/swag/cmd/swag)

tools: build/tools/swag
