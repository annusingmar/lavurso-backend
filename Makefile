include .makerc

.PHONY: api/run
api/run:
	@clear
	@go run ./cmd/api

.PHONY: db/migration_up
db/migration_up:
	@migrate -path=./migrations -database="${MIGRATE_DSN}" up

.PHONY: db/migration_down
db/migration_down:
	@migrate -path=./migrations -database="${MIGRATE_DSN}" down

.PHONY: db/migration_create
db/migration_create:
	@migrate create -ext=.sql -dir=./migrations -seq ${NAME}