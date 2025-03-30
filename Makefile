SHELL := /usr/bin/env bash

.PHONY: all
all: build

.PHONY: clean
clean:
	rm bin/*

.PHONY: test
test:
	go test -cover -timeout 60s -v ./...

bin/server:
	go build -o bin/server cmd/server/main.go

bin/client:
	go build -o bin/client main.go

.PHONY: build
build: bin/server bin/client

SERVER_ARGS := --log-level info --scheme http

.PHONY: server
server: bin/server
	./bin/server $(SERVER_ARGS) | jq .

GH_KEY ?= ~/.ssh/id_ecdsa
GH_USERNAME ?= $(shell whoami)

CLIENT_ARGS ?= 

.PHONY: client
client: bin/client
	echo "Set GH_KEY to the path of your private key registered with GitHub"
	echo "Set GH_USERNAME to your GitHub username"
	./bin/client --key $(GH_KEY) --username $(GH_USERNAME) $(CLIENT_ARGS)

# Docker build variables
DOCKER_IMAGE ?= 125843596666.dkr.ecr.us-west-2.amazonaws.com/rejekts/httpsig-server
DOCKER_TAG ?= latest

.PHONY: docker-build
docker-build:
	docker buildx create --name rejekts-builder --use || true
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--tag $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--load \
		.

docker-push:
	docker buildx create --name rejekts-builder --use || true
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--tag $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--push \
		.
