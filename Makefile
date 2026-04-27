# Makefile bootstrap + Taskfile forwarder.

TASK_PACKAGE := github.com/go-task/task/v3/cmd/task@latest
GOBIN := $(shell if command -v go >/dev/null 2>&1; then gobin=$$(go env GOBIN); if [ -n "$$gobin" ]; then printf '%s' "$$gobin"; else printf '%s' "$$(go env GOPATH)/bin"; fi; fi)
TASK := PATH="$(if $(GOBIN),$(GOBIN):)$(PATH)" task
UI_DIRS := testrunner/ui pr/ui

.PHONY: build lint test install restart fmt tidy clean all ensure-task deps

ensure-task:
	@if ! PATH="$(if $(GOBIN),$(GOBIN):)$(PATH)" command -v task >/dev/null 2>&1; then \
		echo "Installing task ($(TASK_PACKAGE))"; \
		command -v go >/dev/null 2>&1 || { echo "go is required to install task"; exit 1; }; \
		GOBIN="$(GOBIN)" go install $(TASK_PACKAGE); \
	fi

deps: ensure-task
	@echo "Ensuring Go module dependencies are available"
	@go mod download
	@for dir in $(UI_DIRS); do \
		if [ ! -d "$$dir/node_modules" ]; then \
			echo "Installing $$dir dependencies"; \
			command -v npm >/dev/null 2>&1 || { echo "npm is required to install $$dir dependencies"; exit 1; }; \
			( cd "$$dir" && npm ci ); \
		fi; \
	done

build: deps
	@$(TASK) build

lint: deps
	@$(TASK) lint

test: deps
	@$(TASK) test

install: deps
	@$(TASK) install

restart: deps
	@$(TASK) restart

fmt: ensure-task
	@$(TASK) fmt

tidy: ensure-task
	@$(TASK) mod

clean: ensure-task
	@$(TASK) clean

all: deps
	@$(TASK) ci
