.PHONY: run
run: run-api

.PHONY: run-api
run-api:
	@go run cmd/core/main.go

.PHONY: test
test:
	@bash -c '\
		trap "docker stop test-mongodb && docker rm test-mongodb" EXIT; \
		docker run -d --name test-mongodb -p 27018:27017 -e MONGO_INITDB_ROOT_USERNAME=root -e MONGO_INITDB_ROOT_PASSWORD=password mongo:latest; \
		sleep 2s; \
		go test ./... -v -db-host=localhost -db-port=27018 -db-user=root -db-password=password -db-name=bank; \
	'
