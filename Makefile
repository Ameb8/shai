BINARY       = bin/_shai_bin
BINARY_DEBUG = bin/_shai_bin_dbg

.PHONY: all build debug test fmt lint clean tidy doc help

all: build debug ## Build release and debug binaries


## --- --- Build --- --- ---
build: ## Compile the release binary
	go build -o $(BINARY) .

debug: ## Compile the debug binary (debug build tag enabled)
	go build -tags debug -o $(BINARY_DEBUG) .


## --- --- Development --- --- --- 

fmt: ## Format all Go source files
	go fmt ./...

tidy: ## Tidy and verify go.mod / go.sum
	go mod tidy

doc: ## Install pkgsite and serve docs locally
	go install golang.org/x/pkgsite/cmd/pkgsite@latest
	pkgsite -open .


## --- --- Quality --- --- --- 

test: ## Run the test suite
	go test ./...


## --- --- Maintenance --- --- ---

clean: ## Remove build artifacts
	rm -rf bin/


## --- --- Help --- --- ---

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'