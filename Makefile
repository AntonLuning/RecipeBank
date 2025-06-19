MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MAKEFILE_DIR := $(shell dirname $(MAKEFILE_PATH))

BIN_PATH := "$(MAKEFILE_DIR)/bin"
ASSETS_PATH := "$(BIN_PATH)/assets"

MONGO_PASSWORD := my_mongo_password

.PHONY: run-api
run-api: build-api
	@mkdir -p $(BIN_PATH)
	@echo -n $(MONGO_PASSWORD) > $(BIN_PATH)/db_password
	@export \
		RP_DB_HOST="localhost" \
		RP_DB_USERNAME="mongoadmin" \
		RP_DB_PASSWORD_FILE="$(BIN_PATH)/db_password" \
		RP_DB_DATABASE="recipes_db" \
		RP_AI_PROVIDER="openai" \
		RP_AI_API_KEY=$(shell cat secrets/openai_key) &&\
	$(BIN_PATH)/api

.PHONY: build-api
build-api:
	@go build -o $(BIN_PATH)/api cmd/api/main.go

.PHONY: run-ui
run-ui: build-ui
	@mkdir -p $(BIN_PATH)
	@export \
		RP_UI_DEBUG="true" \
		RP_UI_ASSETS_PATH="$(ASSETS_PATH)" \
		RP_UI_API_URL="http://localhost:9876/api/v1" &&\
	$(BIN_PATH)/ui

.PHONY: build-ui
build-ui: generate-templ generate-assets
	@go build -o $(BIN_PATH)/ui cmd/ui/main.go

.PHONY: generate-templ
generate-templ:
	@templ generate

.PHONY: generate-assets
generate-assets:
	@npx tailwindcss -i ./assets/css/input.css -o $(ASSETS_PATH)/css/output.css --content "./internal/ui/**/*.{templ,go}" --content "./internal/ui/components/**/*.{templ,go}"
	@cp -r ./assets/img/ $(ASSETS_PATH)/.
	@cp -r ./assets/js/ $(ASSETS_PATH)/.

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

.PHONY: test
test:
	@go test ./...

.PHONY: test-ai
test-ai:
	@export \
		OPENAI_API_KEY=$(shell cat secrets/openai_key) \
		TEST_IMAGE_PATH="$(MAKEFILE_DIR)/testdata/recipe_omelett.jpeg" &&\
	go test ./internal/api/ai/...

.PHONY: swagger-docs
swagger-docs:
	@swag init -g internal/api/docs.go -o docs/
