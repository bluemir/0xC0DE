##@ Security

.PHONY: vulncheck
vulncheck: runtime/tools/govulncheck ## Run govulncheck
	govulncheck ./...

.PHONY: sec
sec: runtime/tools/gosec ## Run gosec
	gosec -quiet ./...

runtime/tools/govulncheck:
	GOBIN=$(shell pwd)/runtime/tools go install golang.org/x/vuln/cmd/govulncheck@latest

runtime/tools/gosec:
	GOBIN=$(shell pwd)/runtime/tools go install github.com/securego/gosec/v2/cmd/gosec@latest

build-tools: runtime/tools/govulncheck runtime/tools/gosec
