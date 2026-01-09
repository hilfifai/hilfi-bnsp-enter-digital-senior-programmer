BINARY_NAME=api-jellyfish
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
.DEFAULT_GOAL := run

install:
	go mod download
	go mod verify
	go mod tidy

clean:
	go clean
	rm -rf bin

lint:
	golangci-lint run -v

lint_fix:
	golangci-lint run -v --fix

.PHONY: docs
docs:
	swag init -q -g cmd/api/main.go -o cmd/api/docs

build: lint_fix build_api

build_run: lint_fix build_run_api

run: lint_fix run_api

run_api:
	go run cmd/api/main.go

build_api:
	GOARCH=amd64 GOOS=darwin go build -o bin/api/api-$(BINARY_NAME)-darwin ./cmd/api/main.go
	GOARCH=amd64 GOOS=linux go build -o bin/api/api-$(BINARY_NAME)-linux ./cmd/api/main.go
	cp ./cmd/api/.env ./bin/api/.env

build_run_api: build_api
	cd bin/api && ./api-$(BINARY_NAME)-darwin || cd -

test:
	go test ./... -v

test_coverage:
	go test ./... -v -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

benchmark:
	go test ./... -bench=. -benchmem