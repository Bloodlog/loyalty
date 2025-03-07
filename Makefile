GOLANGCI_LINT_CACHE?=/tmp/praktikum-golangci-lint-cache

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.57.2 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint

style:
	go mod tidy -compat=1.22
	go fmt ./...
	go vet ./...
	goimports -w .

build:
	docker-compose up --build

migrate:
	docker run --rm \
    -v $(realpath ./internal/db/migrations):/migrations \
    --network gophermart \
    migrate/migrate:v4.18.1 \
        -path=/migrations \
        -database postgresql://gopher:gopher@postgres:5432/gophermart?sslmode=disable \
        up

 migrate-rollback:
	docker run --rm \
	    -v $(realpath ./internal/db/migrations):/migrations \
	    --network gophermart \
	    migrate/migrate:v4.18.1 \
	        -path=/migrations \
	        -database postgresql://gopher:gopher@postgres:5432/gophermart?sslmode=disable \
	        down -all

mock:
	mockgen -source=internal/app/repositories/order_repository.go \
		-destination=internal/app/repositories/mocks/order_repository_mock.go \
		-package=mocks
	mockgen -source=internal/app/repositories/user_repository.go \
		-destination=internal/app/repositories/mocks/user_repository_mock.go \
		-package=mocks
	mockgen -source=internal/app/repositories/job_repository.go \
		-destination=internal/app/repositories/mocks/job_repository_mock.go \
		-package=mocks
	mockgen -source=internal/app/repositories/withdraw_repository.go \
		-destination=internal/app/repositories/mocks/withdraw_repository_mock.go \
		-package=mocks

go-test:
	go test ./...

go-test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out