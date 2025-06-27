MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MAKEFILE_DIR := $(shell dirname $(MAKEFILE_PATH))

BIN_PATH := "$(MAKEFILE_DIR)/bin"
ASSETS_PATH := "$(BIN_PATH)/assets"

.PHONY: run-api
run-api: build-api
	@mkdir -p $(BIN_PATH)
	@cp $(MAKEFILE_DIR)/secrets/mongo_password $(BIN_PATH)/db_password
	@cp $(MAKEFILE_DIR)/secrets/openai_key $(BIN_PATH)/openai_key
	@export \
		RP_DB_HOST="localhost" \
		RP_DB_PORT="27017" \
		RP_DB_USERNAME="mongoadmin" \
		RP_DB_PASSWORD_FILE="$(BIN_PATH)/db_password" \
		RP_DB_DATABASE="recipes_db" \
		RP_AI_API_KEY_FILE="$(BIN_PATH)/openai_key" &&\
	$(BIN_PATH)/api

.PHONY: build-api
build-api:
	@go build -o $(BIN_PATH)/api $(MAKEFILE_DIR)/cmd/api/main.go

.PHONY: build-api-docker
build-api-docker:
	@docker build -t recipe-bank-api:latest -f $(MAKEFILE_DIR)/cmd/api/Dockerfile .

.PHONY: run-ui
run-ui: build-ui
	@mkdir -p $(BIN_PATH)
	@export \
		RP_UI_DEBUG="true" \
		RP_UI_ASSETS_PATH="$(ASSETS_PATH)" \
		RP_UI_API_URL="http://localhost:9876" \
		RP_UI_API_BASE_PATH="/api/v1" &&\
	$(BIN_PATH)/ui

.PHONY: build-ui
build-ui: generate-templ generate-assets
	@go build -o $(BIN_PATH)/ui $(MAKEFILE_DIR)/cmd/ui/main.go

.PHONY: build-ui-docker
build-ui-docker: generate-templ generate-assets
	@docker build -t recipe-bank-ui:latest -f $(MAKEFILE_DIR)/cmd/ui/Dockerfile .

.PHONY: generate-templ
generate-templ:
	@templ generate

.PHONY: generate-assets
generate-assets:
	@npx tailwindcss -i $(MAKEFILE_DIR)/assets/css/input.css -o $(ASSETS_PATH)/css/output.css \
	   --content "$(MAKEFILE_DIR)/internal/ui/**/*.{templ,go}" \
	   --content "$(MAKEFILE_DIR)/internal/ui/components/**/*.{templ,go}"
	@cp -r $(MAKEFILE_DIR)/assets/img/ $(ASSETS_PATH)/.
	@cp -r $(MAKEFILE_DIR)/assets/js/ $(ASSETS_PATH)/.

.PHONY: mongo-start
mongo-start:
	@docker run -d --name mongodb-recipebank \
		-e MONGO_INITDB_ROOT_USERNAME="mongoadmin" \
		-e MONGO_INITDB_ROOT_PASSWORD=$(shell cat $(MAKEFILE_DIR)/secrets/mongo_password) \
		-p 27017:27017 \
		mongo:latest

.PHONY: mongo-stop
mongo-stop:
	@docker stop mongodb-recipebank
	@docker rm mongodb-recipebank

.PHONY: docker-compose-up
docker-compose-up:
	@docker compose -f $(MAKEFILE_DIR)/examples/docker-compose.yml up -d

.PHONY: docker-compose-down
docker-compose-down:
	@docker compose -f $(MAKEFILE_DIR)/examples/docker-compose.yml down

.PHONY: test
test:
	@go test $(MAKEFILE_DIR)/...

.PHONY: test-ai
test-ai:
	@export \
		OPENAI_API_KEY=$(shell cat secrets/openai_key) \
		TEST_IMAGE_PATH="$(MAKEFILE_DIR)/testdata/recipe_omelett.jpeg" &&\
	go test $(MAKEFILE_DIR)/internal/api/ai/...

.PHONY: swagger-docs
swagger-docs:
	@swag init -g $(MAKEFILE_DIR)/internal/api/docs.go -o docs/
