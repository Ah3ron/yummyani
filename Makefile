BINARY   := yummyani
VERSION  := 2.0.0
SRC      := ./cmd/yummyani
PREFIX   ?= /usr/local

.PHONY: build run clean install uninstall lint fmt vet test archive help

build: ## Build binary
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY) $(SRC)

run: build ## Build and run
	./$(BINARY)

clean: ## Remove binary and archive
	rm -f $(BINARY) $(BINARY).tar.gz

install: build ## Install to PREFIX (default /usr/local)
	install -Dm755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

uninstall: ## Remove installed binary
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY)

lint: vet ## Run all linters (vet only — extend with staticcheck)

vet: ## Run go vet
	go vet ./...

fmt: ## Format code with gofmt
	gofmt -w .
	@if command -v gofumpt >/dev/null 2>&1; then gofumpt -w .; fi

test: ## Run all tests with race detector
	go test -v -race -count=1 ./...

cover: ## Run tests and generate coverage report
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

archive: build ## Create tar.gz archive
	tar czf ./$(BINARY).tar.gz -C .. \
		--exclude=$(BINARY).tar.gz \
		--exclude=$(BINARY)/$(BINARY) \
		--exclude=.git \
		--exclude=vendor \
		$(BINARY)/
	@echo "Created $(BINARY).tar.gz"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
