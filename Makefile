APP := follower
CMD_DIR := ./cmd/$(APP)
BIN_DIR := ./bin
BIN := $(BIN_DIR)/$(APP)
GOFMT_FILES := $(shell rg --files -g'*.go' .)
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
INSTALL ?= install

.PHONY: help fmt build install uninstall run test tidy clean check

help:
	@printf "Available targets:\n"
	@printf "  make fmt      - format Go files\n"
	@printf "  make build    - build $(APP) into $(BIN)\n"
	@printf "  make install  - install $(APP) into $(DESTDIR)$(BINDIR)\n"
	@printf "  make uninstall - remove $(APP) from $(DESTDIR)$(BINDIR)\n"
	@printf "  make run      - run $(APP) from source\n"
	@printf "  make test     - run tests\n"
	@printf "  make tidy     - tidy Go modules\n"
	@printf "  make clean    - remove built artifacts\n"
	@printf "  make check    - run fmt, test, and tidy\n"

fmt:
	gofmt -w $(GOFMT_FILES)

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN) $(CMD_DIR)

install:
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

clean:
	rm -rf $(BIN_DIR)

check: fmt test tidy
