##@ Run

run: build/$(APP_NAME) runtime/certs/server.crt runtime/certs/server.key ## Run web app
	$< -vvv server --cert runtime/certs/server.crt --key runtime/certs/server.key
dev-run: ## Run dev server. If detect file change, automatically rebuild&restart server
	@$(MAKE) build/tools/watcher
	watcher \
		--include "go.mod" \
		--include "go.sum" \
		--include "**.go" \
		--include "package.json" \
		--include "yarn.lock" \
		--include "web/**" \
		--include "api/proto/**" \
		--include "Makefile" \
		--include "scripts/makefile.d/*.mk" \
		--exclude "build/**" \
		-- \
	$(MAKE) test run


reset: ## Kill all make process. Use when dev-run stuck.
	ps -e | grep make | grep -v grep | awk '{print $$1}' | xargs kill

tools: build/tools/watcher
build/tools/watcher: build/tools/go
	@which $(notdir $@) || (./scripts/tools/install-go-tool.sh github.com/bluemir/watcher)

.PHONY: run dev-run reset
