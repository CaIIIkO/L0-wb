include .env
export $(shell sed 's/=.*//' .env)

run:
	go run cmd/main.go

#============Docker============
# Запустить контейнеры
up:
	docker-compose up -d

# Остановить контейнеры
down:
	docker-compose down

# Перезапустить контейнер (с пересборкой и пересозданием)
restart: down up

#============МИГРАЦИИ============
goose-install:
	go install github.com/pressly/goose/v3/cmd/goose@latest

goose-add:
	goose -dir ./migrations postgres "$(DATABASE_DSN_MIGRATIONS)" create rename_me sql

goose-up:
	goose -dir ./migrations postgres "$(DATABASE_DSN_MIGRATIONS)" up

goose-down:
	goose -dir ./migrations postgres "$(DATABASE_DSN_MIGRATIONS)" down

goose-status:
	goose -dir ./migrations postgres "$(DATABASE_DSN_MIGRATIONS)" status