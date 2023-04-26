.PHONY: test
test:
	@go test -race -cover $(shell go list ./... | grep -v -E '')

.PHONY: check check/lint check/gosec check/spell check/trailing
check: check/lint check/gosec check/spell check/trailing

check/lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

check/gosec:
	@echo "Running gosec..."
	@gosec -quiet -exclude-generated ./...

check/spell:
	@echo "Verifying spelling..."
	@yarn cspell --no-progress '**/**'

check/trailing:
	@echo "Looking for trailing whitespaces..."
	@yarn check-trailing-whitespaces
