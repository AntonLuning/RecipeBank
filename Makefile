MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MAKEFILE_DIR := $(shell dirname $(MAKEFILE_PATH))

BIN_PATH := "$(MAKEFILE_DIR)/bin"
ASSETS_PATH := "$(BIN_PATH)/assets"

MONGO_PASSWORD := my_mongo_password

.PHONY: run-core
run-core: build-core
	@mkdir -p $(BIN_PATH)
	@echo -n $(MONGO_PASSWORD) > $(BIN_PATH)/db_password
	@export \
		RP_DB_HOST="localhost" \
		RP_DB_USERNAME="mongoadmin" \
		RP_DB_PASSWORD_FILE="$(BIN_PATH)/db_password" \
		RP_DB_DATABASE="recipes_db" \
		RP_OPENAI_API_KEY=$(shell cat secrets/openai_key) &&\
	./$(BIN_PATH)/core

.PHONY: build-core
build-core:
	@go build -o $(BIN_PATH)/core cmd/core/main.go

.PHONY: run-ui
run-ui: setup-ui-assets build-ui
	@mkdir -p $(BIN_PATH)
	@export \
		RP_UI_DEBUG="true" \
		RP_UI_ASSETS_PATH="$(ASSETS_PATH)" &&\
	./$(BIN_PATH)/ui

.PHONY: build-ui
build-ui:
	@go build -o $(BIN_PATH)/ui cmd/ui/main.go

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

.PHONY: setup-ui-assets
setup-ui-assets: compile_tailwind generate_templ
	@mkdir -p $(ASSETS_PATH)/js
	@find $(MAKEFILE_DIR)/3rd -type f \( -name "*.js" \) | \
	while IFS= read -r file; do \
		cp "$$file" $(ASSETS_PATH)/js/; \
	done
	@mkdir -p $(ASSETS_PATH)/img

.PHONY: compile_tailwind
compile_tailwind:
	@cd $(MAKEFILE_DIR)/tailwind && npx tailwindcss -i $(MAKEFILE_DIR)/internal/ui/views/static/css/input.css -o $(ASSETS_PATH)/css/output.css

.PHONY: generate_templ
generate_templ:
	@templ generate -path $(MAKEFILE_DIR)/internal/ui/views/

.PHONY: test
test:
	@go test ./...

.PHONY: test-ai
test-ai:
	@export \
		OPENAI_API_KEY=$(shell cat secrets/openai_key) \
		TEST_IMAGE_PATH="$(MAKEFILE_DIR)/testdata/test_image.jpg" &&\
	go test ./internal/core/ai/...

