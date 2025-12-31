DB_URL=postgres://user:pass@auth-db:5432/gatekeeper_db?sslmode=disable
NETWORK_NAME=gatekeeper-network

MIGRATE_IMAGE = migrate/migrate
MIGRATION_DIR = db/migration

postgres:
	docker compose up -d

migrateup:
	docker run --rm -v $(shell pwd)/$(MIGRATION_DIR):/migrations --network $(NETWORK_NAME) $(MIGRATE_IMAGE) -path=/migrations -database "$(DB_URL)" up

migratedown:
	docker run --rm -v $(shell pwd)/$(MIGRATION_DIR):/migrations --network $(NETWORK_NAME) $(MIGRATE_IMAGE) -path=/migrations -database "$(DB_URL)" down 1

new_migration:
	docker run --rm -v $(shell pwd)/$(MIGRATION_DIR):/migrations $(MIGRATE_IMAGE) create -ext sql -dir /migrations -seq $(name)

migrate_force:
	docker run --rm -v $(shell pwd)/$(MIGRATION_DIR):/migrations --network $(NETWORK_NAME) $(MIGRATE_IMAGE) -path=/migrations -database "$(DB_URL)" force $(version)

.PHONY: postgres migrateup migratedown new_migration migrate_force