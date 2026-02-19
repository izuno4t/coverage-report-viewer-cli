APP := jrv
CMD := ./cmd/jrv
OUT := ./bin/$(APP)

.PHONY: help run build test test-perf lint-md fmt tidy verify build-darwin-arm64 build-darwin-amd64 build-linux-amd64

help:
	@echo "Available targets:"
	@echo "  make run                # run locally"
	@echo "  make build              # build ./bin/jrv"
	@echo "  make test               # run all tests"
	@echo "  make test-perf          # run JaCoCo parser perf test"
	@echo "  make lint-md            # lint markdown files"
	@echo "  make fmt                # format Go files"
	@echo "  make tidy               # tidy go.mod/go.sum"
	@echo "  make verify             # test + markdown lint"
	@echo "  make build-darwin-arm64 # cross build"
	@echo "  make build-darwin-amd64 # cross build"
	@echo "  make build-linux-amd64  # cross build"

run:
	go run $(CMD)

build:
	go build -o $(OUT) $(CMD)

test:
	go test ./...

test-perf:
	go test ./internal/jacoco -run TestParsePerformance1000Classes -count=1 -v

lint-md:
	markdownlint-cli2 "**/*.md"

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

verify: test lint-md

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o /tmp/$(APP)-darwin-arm64 $(CMD)

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o /tmp/$(APP)-darwin-amd64 $(CMD)

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o /tmp/$(APP)-linux-amd64 $(CMD)
