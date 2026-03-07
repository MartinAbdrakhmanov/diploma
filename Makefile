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


build:
	go build -o ./bin/app  cmd/cmd/main.go

run: 
	go build -o ./bin/app  cmd/cmd/main.go && ./bin/app



# Команда для очистки реестра
registry-gc:
	docker exec registry bin/registry garbage-collect /etc/docker/registry/config.yml --delete-untagged

# Посмотреть что сейчас в реестре
registry-ls:
	@curl -s http://localhost:5000/v2/_catalog | jq .