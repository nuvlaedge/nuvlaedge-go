
MAIN_PACKAGE_PATH := ./cmd/nuvlaedge/
BINARY_NAME := nuvlaedge

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go test -race -buildvcs -vet=off ./...

## lint: Run linter on the Go source code
.PHONY: lint
lint:
	golangci-lint run

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	go test -tags=coverage -v -race -buildvcs $(shell go list ./... | grep -v -e testutils -e cmd/tests)

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -tags=coverage -v -race -buildvcs -coverprofile=/tmp/coverage.out $(shell go list ./... | grep -v -e testutils -e cmd/tests)
	go tool cover -html=/tmp/coverage.out

# docker/build: build docker image
.PHONY: docker/build
docker/build:
	docker build -t local/nuvlaedge:refactor .

# docker/run: builds and runs the docker image using docker compose
.PHONY: docker/run
docker/run:
	docker build -t local/nuvlaedge:refactor .
	docker compose -p nuvlaedge -f docker-compose.yml up

# ==================================================================================== #
# CI/CD
# ==================================================================================== #
.PHONY: ci/test/cover
ci/test/cover:
	go test -tags=coverage -v -race -buildvcs -coverprofile=/coverage.out $(shell go list ./... | grep -v -e testutils -e cmd/tests)


.PHONY: ci/lint
ci/lint:
	golangci-lint run ./...

.PHONY: ci/sec
ci/sec:
	gosec -fmt=sonarqube -out report.json -exclude-dir cmd/tests ./...
