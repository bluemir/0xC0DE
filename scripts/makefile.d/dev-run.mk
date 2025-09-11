##@ Run

run: build/$(APP_NAME) ## Run web app
	$< -vvv server --config runtime/config.yaml #--cert runtime/certs/server.crt --key runtime/certs/server.key
dev-run: | build/tools/watcher ## Run dev server. If detect file change, automatically rebuild&restart server
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
		--exclude "build/**" \
		--exclude "**.sw*" \
		--exclude "assets/js/index.js" \
		--exclude "pkg/api/v1/**.go" \
		-- \
	$(MAKE) test run

test-run: | build/tools/watcher ## Run test. If detect file change, automatically run test
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
		--exclude "build/**" \
		--exclude "**.sw*" \
		--exclude "assets/js/index.js" \
		-- \
	$(MAKE) test

reset: ## Kill all make process. Use when dev-run stuck.
	ps -e | grep $(APP_NAME) | grep -v grep | awk '{print $$1}' | xargs kill

tools: build/tools/watcher
build/tools/watcher: build/tools/go
	@which $(notdir $@) || (./scripts/tools/install/go-tool.sh github.com/bluemir/watcher)

.PHONY: run dev-run reset
