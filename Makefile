APP := follower
CMD_DIR := ./cmd/$(APP)
BIN_DIR := ./bin
BIN := $(BIN_DIR)/$(APP)
GOFMT_FILES := $(shell rg --files -g'*.go' .)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || printf unknown)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
INSTALL ?= install

.PHONY: help fmt fmt-check build install uninstall run test tidy tidy-check clean check ci

help:
	@printf "Available targets:\n"
	@printf "  make fmt      - format Go files\n"
	@printf "  make fmt-check - verify Go files are formatted\n"
	@printf "  make build    - build $(APP) into $(BIN)\n"
	@printf "  make install  - install $(APP) into $(DESTDIR)$(BINDIR)\n"
	@printf "  make uninstall - remove $(APP) from $(DESTDIR)$(BINDIR)\n"
	@printf "  make run      - run $(APP) from source\n"
	@printf "  make test     - run tests\n"
	@printf "  make tidy     - tidy Go modules\n"
	@printf "  make tidy-check - verify Go modules are tidy\n"
	@printf "  make clean    - remove built artifacts\n"
	@printf "  make check    - run fmt, test, and tidy\n"
	@printf "  make ci       - run non-mutating CI checks and build\n"

fmt:
	gofmt -w $(GOFMT_FILES)

fmt-check:
	@files="$$(gofmt -l $(GOFMT_FILES))"; \
	if [ -n "$$files" ]; then \
		printf "Go files need formatting:\n%s\n" "$$files"; \
		exit 1; \
	fi

build:
	mkdir -p $(BIN_DIR)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD_DIR)

install: build
	$(INSTALL) -d $(DESTDIR)$(BINDIR)
	$(INSTALL) -m 0755 $(BIN) $(DESTDIR)$(BINDIR)/$(APP)

uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(APP)

run:
	go run $(CMD_DIR)

test:
	go test ./...

tidy:
	go mod tidy

tidy-check:
	go mod tidy -diff

clean:
	rm -rf $(BIN_DIR)

check: fmt test tidy

ci: fmt-check test tidy-check build
