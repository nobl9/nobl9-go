.DEFAULT_GOAL := help
MAKEFLAGS += --silent --no-print-directory

BIN_DIR := ./bin

# renovate datasource=github-releases depName=abice/go-enum
GO_ENUM_VERSION := v0.6.0
# renovate datasource=github-releases depName=securego/gosec
GOSEC_VERSION := v2.20.0
# renovate datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION := v1.59.1
# renovate datasource=go depName=golang.org/x/vuln/cmd/govulncheck
GOVULNCHECK_VERSION := v1.1.2
# renovate datasource=go depName=golang.org/x/tools/cmd/goimports
GOIMPORTS_VERSION := v0.22.0
# renovate datasource=go depName=github.com/vburenin/ifacemaker
IFACEMAKER_VERSION := v1.2.1

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

# Print Makefile target step description for check.
# Only print 'check' steps this way, and not dependent steps, like 'install'.
# ${1} - step description
define _print_check_step
	printf -- '------\n%s...\n' "${1}"
endef

.PHONY: test test/e2e test/record
## Run all unit tests.
test:
	go test -race -cover ./... ./docs/mock_example

## Run all end-to-end tests (requires Nobl9 platform credentials).
test/e2e:
	go test -race -test.v -timeout=5m -tags=e2e_test ./tests

## Record tests and save them in ./bin/recorded-tests.json.
test/record:
	RECORD_FILE="$(abspath $(dir .))/$(BIN_DIR)/recorded-tests" ; \
	NOBL9_SDK_TEST_RECORD_FILE="$$RECORD_FILE" go test ./... ; \
	jq -s < "$$RECORD_FILE" > "$$RECORD_FILE.json"

.PHONY: check check/vet check/lint check/gosec check/spell check/trailing check/markdown check/format check/generate check/vulns
## Run all checks.
check: check/vet check/lint check/gosec check/spell check/trailing check/markdown check/format check/generate check/vulns

## Run 'go vet' on the whole project.
check/vet:
	$(call _print_check_step,Running go vet)
	go vet ./...

## Run golangci-lint all-in-one linter with configuration defined inside .golangci.yml.
check/lint:
	$(call _print_check_step,Running golangci-lint)
	$(call _ensure_installed,binary,golangci-lint)
	$(BIN_DIR)/golangci-lint run

## Check for security problems using gosec, which inspects the Go code by scanning the AST.
check/gosec:
	$(call _print_check_step,Running gosec)
	$(call _ensure_installed,binary,gosec)
	$(BIN_DIR)/gosec -exclude-generated -quiet ./...

## Check spelling, rules are defined in cspell.json.
check/spell:
	$(call _print_check_step,Verifying spelling)
	$(call _ensure_installed,yarn,cspell)
	yarn --silent cspell --no-progress '**/**'

## Check for trailing whitespaces in any of the projects' files.
check/trailing:
	$(call _print_check_step,Looking for trailing whitespaces)
	yarn --silent check-trailing-whitespaces

## Check markdown files for potential issues with markdownlint.
check/markdown:
	$(call _print_check_step,Verifying Markdown files)
	$(call _ensure_installed,yarn,markdownlint)
	yarn --silent markdownlint '**/*.md' --ignore node_modules

## Check for potential vulnerabilities across all Go dependencies.
check/vulns:
	$(call _print_check_step,Running govulncheck)
	$(call _ensure_installed,binary,govulncheck)
	$(BIN_DIR)/govulncheck ./...

## Verify if the auto generated code has been committed.
check/generate:
	$(call _print_check_step,Checking if generated code matches the provided definitions)
	./scripts/check-generate.sh

## Verify if the files are formatted.
## You must first commit the changes, otherwise it won't detect the diffs.
check/format:
	$(call _print_check_step,Checking if files are formatted)
	./scripts/check-formatting.sh

.PHONY: generate generate/code generate/diagrams
## Auto generate files.
generate: generate/code generate/plantuml

## Generate Golang code.
generate/code:
	echo "Generating Go code..."
	$(call _ensure_installed,binary,go-enum)
	$(call _ensure_installed,binary,ifacemaker)
	go generate ./... ./docs/mock_example
	${MAKE} format/go

PLANTUML_JAR_URL := https://sourceforge.net/projects/plantuml/files/plantuml.jar/download
PLANTUML_JAR :=  $(BIN_DIR)/plantuml.jar
DIAGRAMS_PATH ?= .

## Generate PNG diagrams from PlantUML files.
generate/plantuml: $(PLANTUML_JAR)
	for path in $$(find $(DIAGRAMS_PATH) -name "*.puml" -type f); do \
  		echo "Generating PNG file(s) for $$path"; \
		java -jar $(PLANTUML_JAR) -tpng $$path; \
  	done

# If the plantuml.jar file isn't already present, download it.
$(PLANTUML_JAR):
	echo "Downloading PlantUML JAR..."
	curl -sSfL $(PLANTUML_JAR_URL) -o $(PLANTUML_JAR)

.PHONY: format format/go format/cspell
## Format files.
format: format/go format/cspell

## Format Go files.
format/go:
	echo "Formatting Go files..."
	$(call _ensure_installed,binary,goimports)
	go fmt ./...
	$(BIN_DIR)/goimports -local=github.com/nobl9/nobl9-go -w .

## Format cspell config file.
format/cspell:
	echo "Formatting cspell.yaml configuration (words list)..."
	$(call _ensure_installed,yarn,yaml)
	yarn --silent format-cspell-config

.PHONY: install install/yarn install/go-enum install/golangci-lint install/gosec install/govulncheck install/goimports install/ifacemaker
## Install all dev dependencies.
install: install/yarn install/go-enum install/golangci-lint install/gosec install/govulncheck install/goimports install/ifacemaker

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

## Install ifacemaker (https://github.com/vburenin/ifacemaker).
install/ifacemaker:
	echo "Installing ifacemaker..."
	$(call _install_go_binary,github.com/vburenin/ifacemaker@$(IFACEMAKER_VERSION))

.PHONY: help
## Print this help message.
help:
	./scripts/makefile-help.awk $(MAKEFILE_LIST)
