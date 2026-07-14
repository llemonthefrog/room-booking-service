DB_CONTAINER=db
DB_USER=user
DB_NAME=booking

.PHONY: up down seed restart test cover

up:
	docker-compose up -d --build

down:
	docker-compose down

restart: down up

seed:
	@echo "Seeding database..."
	docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) < seed.sql
	@echo "Seeding completed."
