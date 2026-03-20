.PHONY: help fmt test tidy check

help:
	@printf "Available targets:\n"
	@printf "  make fmt    - format Go files\n"
	@printf "  make test   - run tests\n"
	@printf "  make tidy   - tidy Go modules\n"
	@printf "  make check  - run fmt, test, and tidy\n"

fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path './.git/*')

test:
	go test ./...

tidy:
	go mod tidy

check: fmt test tidy
