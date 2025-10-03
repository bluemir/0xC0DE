##@ Run

run: build/$(APP_NAME) ## Run web app
	$< -vvv server --config runtime/config.yaml #--cert runtime/certs/server.crt --key runtime/certs/server.key
dev-run: | runtime/tools/watcher ## Run dev server. If detect file change, automatically rebuild&restart server
	watcher \
		--include "go.mod" \
		--include "go.sum" \
		--include "**.go" \
		--include "package.json" \
		--include "yarn.lock" \
		--include "assets/**" \
		--include "api/proto/**" \
		--include "Makefile" \
		--include "scripts/makefile.d/*.mk" \
		--include "runtime/config.yaml" \
		--exclude "build/**" \
		--exclude "**.sw*" \
		--exclude "assets/js/**" \
		--exclude "assets/css/**" \
		--exclude "assets/src/js/index.js" \
		--exclude "pkg/api/v1/**.go" \
		-- \
	$(MAKE) test run

reset: ## Kill all make process. Use when dev-run stuck.
	ps -e | grep $(APP_NAME) | grep -v grep | awk '{print $$1}' | xargs kill

tools: runtime/tools/watcher
runtime/tools/watcher: runtime/tools/go
	@which $(notdir $@) || (./scripts/tools/install/go-tool.sh github.com/bluemir/watcher)

.PHONY: run dev-run reset
