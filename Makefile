# Load environment variables from .env
include .env
export

GOOSE = goose
MIGRATIONS_DIR = ./migrations

DB_URL = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSLMODE)

COMPOSE = docker compose -f deployments/docker/docker-compose.yml

.PHONY: \
	up down restart logs \
	goose-up goose-down goose-status goose-reset goose-create

# -----------------------------------------------------------------------------
# Docker
# -----------------------------------------------------------------------------

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

restart: down up

logs:
	$(COMPOSE) logs -f

# -----------------------------------------------------------------------------
# Goose
# -----------------------------------------------------------------------------

goose-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

goose-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

goose-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

goose-reset:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" reset

goose-create:
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(NAME) sql
