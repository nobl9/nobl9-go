.PHONY: test
test:
	@go test -race -cover ./...

.PHONY: generate
generate:
	@echo "Generating Go code..."
	@go generate ./...

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