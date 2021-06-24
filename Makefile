SHELL := bash
.ONESHELL:
MAKEFLAGS += --no-builtin-rules

.PHONY: help lint

export VERSION := $(if $(TAG),$(TAG),$(if $(BRANCH_NAME),$(BRANCH_NAME),$(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)))
export DOCKER_BUILDKIT := 1

NOCACHE := $(if $(NOCACHE),"--no-cache")

help: ## List all available targets with help
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

lint:
	@golangci-lint run

build-dev:
	@docker build ${NOCACHE} --pull -f ./build/acceptor.Dockerfile -t acceptor:latest .
	@docker build ${NOCACHE} --pull -f ./build/sender.Dockerfile -t sender:latest .

run-dev-infra:
	@docker-compose up -d mongodb rabbitmq

stop-dev: ## Stop develop environment
	@docker-compose down