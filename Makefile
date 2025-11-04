PHONY: install-golangci-lint lint

GOLANGCI_LINT_VERSION=v2.6.1

# Install golangci-lint using the official installer script
install-golangci-lint:
	@if ! command -v golangci-lint > /dev/null || [ "$$(golangci-lint --version 2>/dev/null | grep -oE 'version [0-9]+\.[0-9]+\.[0-9]+' | sed 's/version /v/')" != "${GOLANGCI_LINT_VERSION}" ]; then \
		echo "Installing golangci-lint ${GOLANGCI_LINT_VERSION}..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin ${GOLANGCI_LINT_VERSION}; \
	else \
		echo "golangci-lint ${GOLANGCI_LINT_VERSION} is already installed"; \
	fi

lint: install-golangci-lint
	$$(go env GOPATH)/bin/golangci-lint run -v
