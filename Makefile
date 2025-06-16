include .env
export

DOCKER_NETWORK=go-test_default

# Переменные PostgreSQL
PG_DSN=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
PG_MIGRATION_HOST_PATH=$(shell pwd)/migrations/postgresql

# Переменные ClickHouse
CH_MIGRATIONS=./migrations/clickhouse

# PostgreSQL миграции

migrate-up-pg:
	docker run --rm \
		--network $(DOCKER_NETWORK) \
		-v "/d/gorepos/src/go-test/migrations/postgresql:/migrations" \
		-w /migrations \
		migrate/migrate:latest \
		-source file:///migrations \
		-database "$(PG_DSN)" \
		up

migrate-down-pg:
	docker run --rm \
		--network $(DOCKER_NETWORK) \
		-v "/d/gorepos/src/go-test/migrations/postgresql:/migrations" \
		-w /migrations \
		migrate/migrate:latest \
		-source file:///migrations \
		-database "$(PG_DSN)" \
		down

# ClickHouse миграции
migrate-up-ch:
	find $(CH_MIGRATIONS) -name '*.up.sql' | sort | xargs cat | docker exec -i clickhouse clickhouse-client --multiquery

migrate-down-ch:
	find $(CH_MIGRATIONS) -name '*.down.sql' | sort -r | xargs cat | docker exec -i clickhouse clickhouse-client --multiquery

# Проверка состояния
status-pg:
	docker exec -it $(PG_CONTAINER) psql -U $(PG_USER) -d $(PG_DB) -c '\dt'

status-ch:
	docker exec -it clickhouse clickhouse-client --query "SHOW TABLES"
