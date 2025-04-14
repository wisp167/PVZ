API_PKG := api
SWAGGER_FILE := schema/swagger.yaml
GEN_DIR := api
MODELS_FILE := $(GEN_DIR)/models.gen.go
SERVER_FILE := $(GEN_DIR)/server.gen.go
CLIENT_FILE := $(GEN_DIR)/client.gen.go
SPEC_FILE := $(GEN_DIR)/spec.gen.go
#SQLC_DIR := internal/sql/queries
#SQLC_YAML := sqlc.yaml # Path to your sqlc.yaml file

# Tools
OAPI_CODEGEN := oapi-codegen
GOIMPORTS := goimports

.PHONY: all generate clean fmt lint test help

all: generate fmt# sqlc ## Generate code and format (default target)

gen: $(MODELS_FILE) $(SERVER_FILE) $(CLIENT_FILE) $(SPEC_FILE) ## Generate all code from OpenAPI spec

$(MODELS_FILE): $(SWAGGER_FILE)
	@echo "Generating models..."
	@$(OAPI_CODEGEN) -generate types -package $(API_PKG) $< > $@

$(SERVER_FILE): $(SWAGGER_FILE)
	@echo "Generating server interfaces..."
	@$(OAPI_CODEGEN) -generate server -package $(API_PKG) $< > $@

$(CLIENT_FILE): $(SWAGGER_FILE)
	@echo "Generating client code..."
	@$(OAPI_CODEGEN) -generate client -package $(API_PKG) $< > $@

$(SPEC_FILE): $(SWAGGER_FILE)
	@echo "Generating embedded spec..."
	@$(OAPI_CODEGEN) -generate spec -package $(API_PKG) $< > $@

#sqlc:
#	@echo "Generating sqlc code..."
#	@sqlc generate

fmt: ## Format generated code
	@echo "Formatting generated files..."
	@$(GOIMPORTS) -w $(GEN_DIR)/*.gen.go

clean: ## Remove all generated files
	@echo "Cleaning generated files..."
	@rm -f $(GEN_DIR)/*.gen.go

lint: ## Lint the generated code (add your linter commands here)
	@echo "Linting code..."
	@golangci-lint run

test: ## Run tests
	@echo "Running tests..."
	@go test ./...
