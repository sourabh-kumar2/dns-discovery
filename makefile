.PHONY: fmt vet lint test imports

# Check if goimports is installed
GOIMPORTS := $(shell command -v goimports 2> /dev/null)
# Check if golangci-lint is installed
GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)


lint:
ifeq ($(strip $(GOLANGCI_LINT)),)
	$(error "golangci-lint is not installed. Please install it.")
endif
	$(MAKE) fmt
	$(GOLANGCI_LINT) run

fmt:
	gofmt -s -w .
	$(MAKE) imports

imports:
ifeq ($(strip $(GOIMPORTS)),)
	$(error "goimports is not installed. Please install it using: go get golang.org/x/tools/cmd/goimports")
endif
	$(GOIMPORTS) -w .

vet:
	go vet ./...

test:
	go test -v -cover ./...

# Optional: clean command to clean build artifacts and caches
clean:
	go clean -cache -modcache -i -r
