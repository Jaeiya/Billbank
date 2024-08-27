
GO_INSTALLED := $(shell command -v go version)
GO_VERSION := 1.22
GORELEASER_EXISTS := $(shell command -v goreleaser version)
INSTALL_GORELEASER := go install github.com/goreleaser/goreleaser/v2@latest

# Linux setup
GO_TARBALL := https://golang.org/dl/go$(GO_VERSION).linux-amd64.tar.gz
LINUX_INSTALL_DIR := /usr/local

# Try to install Go on windows if it doesn't exist
install-go-windows:
	@if [ -z "$(GO_INSTALLED)" ]; then \
		echo "Go is not installed. Installing now..."; \
		choco install golang --version $(GO_VERSION) -y; \
	    if [ $$? -ne 0 ]; then \
	        echo "Failed to install Golang. Please install it manually."; \
	        exit 1; \
		fi \
	fi

# Try to install Go on linux if it doesn't exist (partially tested)
install-go-linux:
	@if [ -z "$(GO_INSTALLED)" ]; then \
		echo "Go not found. Installing Go $(GO_VERSION)..."; \
		curl -L $(GO_TARBALL) | sudo tar -C $(LINUX_INSTALL_DIR) -xzf; \
	    if [ $$? -ne 0 ]; then \
	        echo "Failed to install Golang. Please install it manually."; \
	        exit 1; \
		fi \
	fi

# Try to install go on macos if it doesn't exist (not tested)
install-go-macos:
	@if [ -z "$(GO_INSTALLED)" ]; then \
		echo "Go not found. Installing Go $(GO_VERSION)..."; \
		brew install golang@$(GO_VERSION); \
	    if [ $$? -ne 0 ]; then \
	        echo "Failed to install Golang. Please install it manually."; \
	        exit 1; \
		fi \
	fi


# Check if GoReleaser is installed
install-goreleaser:
	@if [ -z "$(GORELEASER_EXISTS)" ]; then \
	    echo "GoReleaser is not installed. Installing..."; \
		$(INSTALL_GORELEASER); \
	    if [ $$? -ne 0 ]; then \
	        echo "Failed to install GoReleaser. Please install it manually."; \
	        exit 1; \
	    fi \
	fi

setup-windows: install-go-windows install-goreleaser
	@echo "You're are all set!"

setup-linux: install-go-linux install-goreleaser
	@echo You're all set!

setup-macos: install-go-macos install-goreleaser
	@echo You're all set!

dev:
	goreleaser release --snapshot --clean --config .goreleaser.dev.yml

snapshot:
	goreleaser release --snapshot --clean
