MAKEFLAGS += --no-print-directory
GO_ENUM_VERSION := v0.5.6
GO_ENUM_PATH := bin/go-enum

.PHONY: test
test:
	@go test -race -cover ./...

.PHONY: check check/vet check/lint check/gosec check/spell check/trailing check/markdown check/vulns
check: check/vet check/lint check/gosec check/spell check/trailing check/markdown check/vulns

check/vet:
	@echo "Running go vet..."
	@go vet ./...

check/lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

check/gosec:
	@echo "Running gosec..."
	@gosec -exclude-generated ./...

check/spell:
	@echo "Verifying spelling..."
	@yarn cspell --no-progress '**/**'

check/trailing:
	@echo "Looking for trailing whitespaces..."
	@yarn check-trailing-whitespaces

check/markdown:
	@echo "Verifying Mardown files..."
	@yarn markdownlint '*.md' --disable MD010 # MD010 does not handle code blocks well.

check/vulns:
	@echo "Running govulncheck..."
	@govulncheck ./...

check/generate:
	@echo "Checking if generate code matches the provided definitions..."
	@./scripts/check-generate.sh

.PHONY: generate
generate:
	@if [ ! -f $(GO_ENUM_PATH) ]; then ${MAKE} install/go-enum ; fi
	@echo "Generating Go code..."
	@go generate ./...

.PHONY: install install/go-enum install/yarn
install: install/go-enum install/yarn

install/go-enum:
	@echo "Downloading go-enum..."
	@curl -fsSL "https://github.com/abice/go-enum/releases/download/$(GO_ENUM_VERSION)/go-enum_$$(uname -s)_$$(uname -m)" -o $(GO_ENUM_PATH) && chmod +x $(GO_ENUM_PATH)

install/yarn:
	@echo "Installing yarn dependencies"
	@yarn install
