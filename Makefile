BINARY  := yummyani
VERSION := 1.0.0
SRC     := ./cmd/yummyani
PREFIX  ?= /usr/local

.PHONY: build run clean install uninstall lint fmt test archive help

build: ## Build binary
	go build -ldflags "-s -w" -o $(BINARY) $(SRC)

run: build ## Build and run
	./$(BINARY)

clean: ## Remove binary and archive
	rm -f $(BINARY) $(BINARY).tar.gz

install: build ## Install to PREFIX (default /usr/local)
	install -Dm755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

uninstall: ## Remove installed binary
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY)

lint: ## Run vet
	go vet ./...

fmt: ## Format code
	gofmt -w .

test: ## Run tests
	go test -v -race ./...

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
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m            \033[0m \n", $$1, $$2}'
