.PHONY: test
test:
	@go test -race -cover ./...

.PHONY: check check/lint check/gosec check/spell check/trailing check/markdown check/install
check: check/lint check/gosec check/spell check/trailing check/markdown

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
