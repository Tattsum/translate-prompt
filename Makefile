.PHONY: help test lint lint-fix fmt build build-cli build-server build-server-api web-build web-install serve dev clean install-tools deps gofix codegen

BIN_DIR := bin
GO := go
PNPM := pnpm
GOLANGCI_LINT := golangci-lint
BACKEND := ./backend/...

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_.-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install golangci-lint v2 and gqlgen CLI
	@echo "Installing golangci-lint v2..."
	@$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	@$(GO) install github.com/99designs/gqlgen@v0.17.93
	@command -v buf >/dev/null 2>&1 || echo "Install buf: https://buf.build/docs/installation"
	@echo "Done."

codegen: ## Generate GraphQL (gqlgen) and Connect (buf) stubs
	cd backend && gqlgen generate
	cd backend && buf generate

gofix: ## Run go fix twice
	@echo "Running go fix (1/2)..."
	@$(GO) fix $(BACKEND)
	@echo "Running go fix (2/2)..."
	@$(GO) fix $(BACKEND)

fmt: gofix ## Format Go (go fix x2 + gofumpt)
	@echo "Running go fmt..."
	@$(GO) fmt $(BACKEND)
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		cd backend && $(GOLANGCI_LINT) fmt; \
	else \
		echo "golangci-lint not found. Run 'make install-tools'."; \
		exit 1; \
	fi

lint: ## Run golangci-lint and frontend eslint
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		cd backend && $(GOLANGCI_LINT) run ./...; \
	else \
		echo "golangci-lint not found. Run 'make install-tools'."; \
		exit 1; \
	fi
	@cd frontend && $(PNPM) run lint 2>/dev/null || true

test: ## Run Go and frontend tests
	$(GO) test $(BACKEND) -count=1
	cd frontend && $(PNPM) test --run 2>/dev/null || true

build: build-cli build-server ## Build CLI and server

build-cli: ## Build translate-prompt CLI
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/translate-prompt ./backend/cmd/translate-prompt

build-server: web-build ## Build server with embedded SPA
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/server ./backend/cmd/server

build-server-api: ## Build API-only server (no embedded SPA, for production)
	@mkdir -p $(BIN_DIR)
	$(GO) build -tags noembed -o $(BIN_DIR)/server-api ./backend/cmd/server

web-install: ## Install frontend dependencies
	cd frontend && $(PNPM) install

web-build: web-install ## Build frontend (GraphQL codegen + Vite)
	cd frontend && $(PNPM) run build

serve: build-server ## Run API server on :8080
	ENV=dev ./$(BIN_DIR)/server --port 8080

dev: ## Run Vite dev server (proxies /query and Connect)
	cd frontend && $(PNPM) run dev

deps: ## Update Go dependencies
	$(GO) get -u $(BACKEND)
	$(GO) mod tidy

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) frontend/dist frontend/node_modules
