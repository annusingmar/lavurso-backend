include .makerc

.PHONY: run
run:
	@clear
	@go run ./cmd/api -dsn="${DSN}"