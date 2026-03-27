##@ Swagger
#.PRECIOUS: internal/swagger/docs.go

# see https://github.com/swaggo/swag for documents

build/$(APP_NAME): internal/swagger/docs.go
test: internal/swagger/docs.go
vet: internal/swagger/docs.go

OPTIONAL_CLEAN+= internal/swagger/docs.go internal/swagger/swagger.json internal/swagger/swagger.yaml

.PHONY: swagger
swagger: internal/swagger/docs.go ## Make swagger file

internal/swagger/docs.go: $(filter ./internal/server/%.go,$(GO_SOURCES))
	@mkdir -p $(dir $@)
	go tool swag init \
		--generalInfo internal/server/routes.go \
		--parseInternal \
		--output $(dir $@)
	# for dependency, add this option: `--parseDependency`
	# for override swaggo, add this option `--overridesFile .swaggo`
