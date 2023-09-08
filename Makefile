GO ?= go

.PHONY: setup
setup:
	@$(GO) build -o .kitty/.bin/kitty-dev/kitty ./cmd/kitty
	@ln -sf kitty-dev/kitty .kitty/.bin/kitty
	@PATH="$(shell pwd)/.kitty/.bin:${PATH}" kitty install
