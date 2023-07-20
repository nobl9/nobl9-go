.DEFAULT_GOAL := help
MAKEFLAGS += --silent --no-print-directory

BIN_DIR := ./bin

# renovate datasource=github-releases depName=abice/go-enum
GO_ENUM_VERSION := v0.5.5
# renovate datasource=github-releases depName=securego/gosec
GOSEC_VERSION := v2.16.0
# renovate datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION := v1.53.3
# renovate datasource=go depName=golang.org/x/vuln/cmd/govulncheck
GOVULNCHECK_VERSION := v0.1.0
# renovate datasource=go depName=golang.org/x/tools/cmd/goimports
GOIMPORTS_VERSION := v0.11.0

# Check if the program is present in $PATH and install otherwise.
# ${1} - oneOf{binary,yarn}
# ${2} - program name
define _ensure_installed
	LOCAL_BIN_DIR=$(BIN_DIR) ./scripts/ensure_installed.sh "${1}" "${2}"
endef

# Install Go binary using 'go install' with an output directory set via $GOBIN.
# ${1} - repository url
define _install_go_binary
	GOBIN=$(realpath $(BIN_DIR)) go install "${1}"
endef

.PHONY: test
## Run all unit tests.
test:
	go test -race -cover ./...

.PHONY: check check/vet check/lint check/gosec check/spell check/trailing check/markdown check/vulns
## Run all checks.
check: check/vet check/lint check/gosec check/spell check/trailing check/markdown check/vulns

## Run 'go vet' on the whole project.
check/vet:
	echo "Running go vet..."
	go vet ./...

## Run golangci-lint all-in-one linter with configuration defined inside .golangci.yml.
check/lint:
	$(call _ensure_installed,binary,golangci-lint)
	echo "Running golangci-lint..."
	$(BIN_DIR)/golangci-lint run

## Check for security problems using gosec, which inspects the Go code by scanning the AST.
check/gosec:
	$(call _ensure_installed,binary,gosec)
	echo "Running gosec..."
	$(BIN_DIR)/gosec -exclude-generated ./...

## Check spelling, rules are defined in cspell.json.
check/spell:
	$(call _ensure_installed,yarn,cspell)
	echo "Verifying spelling..."
	yarn --silent cspell --no-progress '**/**'

## Check for trailing whitespaces in any of the projects' files.
check/trailing:
	echo "Looking for trailing whitespaces..."
	yarn --silent check-trailing-whitespaces

## Check markdown files for potential issues with mardkownlint.
check/markdown:
	$(call _ensure_installed,yarn,markdownlint)
	echo "Verifying Mardown files..."
	yarn --silent markdownlint '*.md' --disable MD010 # MD010 does not handle code blocks well.

## Check for potential vulnerabilities across all Go dependencies.
check/vulns:
	$(call _ensure_installed,binary,govulncheck)
	echo "Running govulncheck..."
	$(BIN_DIR)/govulncheck ./...

## Verify if the auto generated code has been committed.
check/generate:
	echo "Checking if generated code matches the provided definitions..."
	./scripts/check-generate.sh

## Validate Renovate configuration.
check/renovate:
	$(call _ensure_installed,yarn,renovate)
	echo "Validating Renovate configuration..."
	yarn --silent renovate-config-validator

## Verify if the files are formatted.
## You must first commit the changes, otherwise it won't detect the diffs.
check/format:
	echo "Checking if files are formatted..."
	./scripts/check-formatting.sh

.PHONY: generate
## Auto generate code.
generate:
	$(call _ensure_installed,binary,go-enum)
	echo "Generating Go code..."
	go generate ./...

.PHONY: format format/go format/cspell
## Format files.
format: format/go format/cspell

## Format Go files.
format/go:
	$(call _ensure_installed,binary,goimports)
	go fmt ./...
	$(BIN_DIR)/goimports -local=github.com/nobl9/nobl9-go -w .

## Format cspell config file.
format/cspell:
	$(call _ensure_installed,yarn,yaml)
	yarn --silent format-cspell-config

.PHONY: install install/yarn install/go-enum install/golangci-lint install/gosec install/govulncheck install/goimports
## Install all dev dependencies.
install: install/yarn install/go-enum install/golangci-lint install/gosec install/govulncheck install/goimports

## Install JS dependencies with yarn.
install/yarn:
	echo "Installing yarn dependencies..."
	yarn --silent install

## Install go-enum, an enum generator for Go (https://github.com/abice/go-enum).
install/go-enum:
	echo "Downloading go-enum..."
	curl -fsSL "https://github.com/abice/go-enum/releases/download/$(GO_ENUM_VERSION)/go-enum_$$(uname -s)_$$(uname -m)" \
		-o $(BIN_DIR)/go-enum && chmod +x $(BIN_DIR)/go-enum

## Install golangci-lint (https://golangci-lint.run).
install/golangci-lint:
	echo "Installing golangci-lint..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh |\
 		sh -s -- -b $(BIN_DIR) $(GOLANGCI_LINT_VERSION)

## Install gosec (https://github.com/securego/gosec).
install/gosec:
	echo "Installing gosec..."
	curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh |\
 		sh -s -- -b $(BIN_DIR) $(GOSEC_VERSION)

## Install govulncheck (https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck).
install/govulncheck:
	echo "Installing govulncheck..."
	$(call _install_go_binary,golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION))

## Install goimports (https://pkg.go.dev/golang.org/x/tools/cmd/goimports).
install/goimports:
	echo "Installing goimports..."
	$(call _install_go_binary,golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION))

.PHONY: help
## Print this help message.
help:
	./scripts/makefile-help.awk $(MAKEFILE_LIST)
