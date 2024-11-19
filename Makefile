include .env
MIGRATIONS_PATH = ./migrate/migrations

.PHONY: migration
migration:
	@migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(filter-out $@,$(MAKECMDGOALS))
%:
	@:

.PHONY: migrate-up
migrate-up:
	@migrate -database $(DATABASE_URL) -path $(MIGRATIONS_PATH) up

.PHONY: migrate-down
migrate-down:
	@migrate -database $(DATABASE_URL) -path $(MIGRATIONS_PATH) down $(filter-out $@,$(MAKECMDGOALS))

# If e.g, migration 15 is dirty, you can go back
# to migration 14 with 'force-migration 14'
.PHONY: force-migration
force-migration:
	@migrate -path $(MIGRATIONS_PATH) -database $(DATABASE_URL) force $(filter-out $@,$(MAKECMDGOALS))

.PHONY: seed
seed:
	@go run cmd/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./cmd/sapphire/main.go && swag fmt

.PHONY: test-all
test-all:
	@go test -v ./...
