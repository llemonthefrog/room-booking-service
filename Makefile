DB_CONTAINER=db
DB_USER=user
DB_NAME=booking

.PHONY: up down seed restart test test-e2e cover

up:
	docker compose up -d --build

down:
	docker compose down

restart: down up

seed:
	@echo "Seeding database..."
	docker exec -i $(DB_CONTAINER) psql -v ON_ERROR_STOP=1 -U $(DB_USER) -d $(DB_NAME) < scripts/seed.sql
	@echo "Seeding completed."

test:
	go test -short ./...

test-e2e:
	go test -race -count=1 ./...

cover:
	go test -coverpkg=./... -coverprofile=cp.out ./...
