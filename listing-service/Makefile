include .env
export

export PROJECT_ROOT=$(shell pwd)

.DEFAULT_GOAL := help

env-up: ## env: Запустить окружение проекта
	@docker compose up -d listing-postgres listing-redis

env-down: ## env: Остановить окружение проекта
	@docker compose down listing-postgres listing-redis

env-cleanup: ## env: Очистить окружение проекта
	@read -p "Очистить все volume файлы окружения? [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down listing-postgres listing-redis port-forwarder && \
		rm -rf ${PROJECT_ROOT}/out/pgdata && \
		echo "Файлы окружения очищены"; \
	else \
		echo "Очистка окружения отменена"; \
	fi

env-port-forward: ## env: Открыть порты сервисов окружения
	@docker compose up -d port-forwarder

env-port-close: ## env: Закрыть порты сервисов окружения
	@docker compose down port-forwarder

run-server: ## Запустить сервер локально
	@echo "Starting listing-service..."
	@go run cmd/api/main.go

run-server-deploy: ## Запустить сервер в Docker
	@docker compose up -d --build listing-service

migrate-up: ## PostgreSQL: Накатить миграции
	@make migrate-action action=up

migrate-down: ## PostgreSQL: Откатить миграции
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует параметр action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm listing-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@listing-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

show-logs: ## Показать логи
	@docker compose logs -f

logs-cleanup: ## env: Очистить файлы логов
	@read -p "Очистить все log файлы? [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		rm -rf ${PROJECT_ROOT}/out/logs && \
		echo "Файлы логов очищены"; \
	else \
		echo "Очистка логов отменена"; \
	fi
