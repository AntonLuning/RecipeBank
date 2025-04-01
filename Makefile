MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MAKEFILE_DIR := $(shell dirname $(MAKEFILE_PATH))
BIN_PATH := "$(MAKEFILE_DIR)/bin"

MONGO_PASSWORD := my_mongo_password

.PHONY: run
run: run-api

.PHONY: run-api
run-api:
	@mkdir -p $(BIN_PATH)
	@echo -n $(MONGO_PASSWORD) > $(BIN_PATH)/db_password
	@export \
		RP_DB_HOST="localhost" \
		RP_DB_USERNAME="mongoadmin" \
		RP_DB_PASSWORD_FILE="$(BIN_PATH)/db_password" \
		RP_DB_DATABASE="recipes_db" &&\
	go run cmd/core/main.go

.PHONY: test
test:
	@go test ./...

.PHONY: mongo-start
mongo-start:
	@docker run -d --name mongodb-recipebank \
		-e MONGO_INITDB_ROOT_USERNAME="mongoadmin" \
		-e MONGO_INITDB_ROOT_PASSWORD=$(MONGO_PASSWORD) \
		-p 27017:27017 \
		mongo:latest

.PHONY: mongo-stop
mongo-stop:
	@docker stop mongodb-recipebank
	@docker rm mongodb-recipebank
