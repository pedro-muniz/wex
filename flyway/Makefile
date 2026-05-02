.PHONY: migrate rollback clean validate

DB_URL ?= jdbc:postgresql://localhost:5432/wex
DB_USER ?= postgres
DB_PASSWORD ?= postgres

migrate:
	flyway -url=$(DB_URL) -user=$(DB_USER) -password=$(DB_PASSWORD) -locations=filesystem:sql migrate

rollback:
	flyway -url=$(DB_URL) -user=$(DB_USER) -password=$(DB_PASSWORD) -locations=filesystem:sql undo

clean:
	flyway -url=$(DB_URL) -user=$(DB_USER) -password=$(DB_PASSWORD) -locations=filesystem:sql clean

validate:
	flyway -url=$(DB_URL) -user=$(DB_USER) -password=$(DB_PASSWORD) -locations=filesystem:sql validate
