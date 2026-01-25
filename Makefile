include .env

DATABASE_DSN=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

up: 
	docker compose -f docker-compose.base.yaml up -d

down: 
	docker compose -f docker-compose.base.yaml down

restart: down up

#goose 
goose-add:
	goose -dir ./migrations postgres "$(DATABASE_DSN)" create $(NAME) sql

goose-up:
	goose -dir ./migrations postgres "$(DATABASE_DSN)" up

goose-down:
	goose -dir ./migrations postgres "$(DATABASE_DSN)" down

goose-status:
	goose -dir ./migrations postgres "$(DATABASE_DSN)" status


registry:
	sudo docker run -d --restart=always --name registry -p 5000:5000 \
	-v /opt/registry/data:/var/lib/registry \
	registry:2