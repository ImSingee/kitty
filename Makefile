GO ?= go

.PHONY: setup
setup:
	@$(GO) build -o .kitty/.bin/kitty ./cmd/kitty
	@PATH="$(shell pwd)/.kitty/.bin:${PATH}" kitty install
