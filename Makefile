APP_NAME ?= bigrows
GO ?= go
INPUT_DIR ?= ./dir
INPUT_FILE ?= ./dir/transaksi.csv
PARTS ?= 2
GOCACHE ?= /tmp/bigrows-go-cache
RUN_DIR_ARG := $(word 2,$(MAKECMDGOALS))
RUN_DIR := $(if $(RUN_DIR_ARG),$(RUN_DIR_ARG),$(INPUT_DIR))

export GOCACHE

.PHONY: help tidy test build run-file run-dir clean $(RUN_DIR_ARG)

help:
	@printf "Available targets:\n"
	@printf "  make tidy                         Run go mod tidy\n"
	@printf "  make test                         Run unit tests\n"
	@printf "  make build                        Build ./$(APP_NAME)\n"
	@printf "  make run-file INPUT_FILE=... PARTS=...\n"
	@printf "                                    Split one CSV file directly\n"
	@printf "  make run-dir ./dir                Choose CSV from a folder interactively\n"
	@printf "  make run-dir INPUT_DIR=./dir      Same as above, using a variable\n"
	@printf "  make clean                        Remove binary and *_split folders\n"

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

build:
	$(GO) build -o $(APP_NAME)

run-file:
	$(GO) run main.go -file $(INPUT_FILE) -parts $(PARTS)

run-dir:
	$(GO) run main.go -dir $(RUN_DIR)

clean:
	rm -f $(APP_NAME)
	find . -type d -name '*_split' -prune -exec rm -rf {} +

%:
	@:
