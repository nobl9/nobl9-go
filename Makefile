.PHONY: test
test:
	@go test -race -cover $(shell go list ./... | grep -v -E '')

.PHONY: check check/lint check/gosec check/spell
check: check/lint check/gosec check/spell

check/lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

check/gosec:
	@echo "Running gosec..."
	@gosec -quiet -exclude-generated ./...

check/spell:
	@echo "Running cspell..."
	@yarn cspell --no-progress '**/**'