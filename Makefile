generate-mocks:
	mockgen -source debezium/client.go -destination debezium/client_mock.go -package main DebeziumClient

unit-test: generate-mocks
	go test -v -cover ./...

generate-unit-coverage: generate-mocks
	go test -v -cover ./... -coverprofile=unit_coverage.out
	go tool cover -html=unit_coverage.out -o unit_coverage.html

check-pre-commit:
	pre-commit run --all-files
