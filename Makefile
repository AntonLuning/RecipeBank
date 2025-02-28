.PHONY: run
run: run-api

.PHONY: run-api
run-api:
	@go run cmd/core/main.go

.PHONY: test
test:
	@go test ./...
