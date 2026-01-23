OWNER ?= pipekit
IMAGE_NAME ?= databricks-connector
TAG ?= latest
CLUSTER_NAME ?= k3s-default

.PHONY: all
all: docker-build k3d-import

.PHONY: docs
docs:
	go run cmd/databricks-connector/main.go docs

.PHONY: docker-build
docker-build:
	docker build -t $(OWNER)/$(IMAGE_NAME):$(TAG) .
	docker push $(OWNER)/databricks-connector:latest

.PHONY: k3d-import
k3d-import:
	k3d image import $(IMAGE_NAME):$(TAG) -c $(CLUSTER_NAME)

.PHONY: gomod2nix
gomod2nix:
	docker run --rm -v $$(pwd):/app -w /app nixos/nix nix --extra-experimental-features "nix-command flakes" run github:nix-community/gomod2nix -- --dir .

.PHONY: help
help:
	@echo "Usage:"
	@echo "  make docs            Generate CLI documentation"
	@echo "  make docker-build    Build the docker image"
	@echo "  make k3d-import     Import the docker image into k3d cluster ($(CLUSTER_NAME))"
	@echo "  make gomod2nix      Update gomod2nix.toml using Docker"
	@echo "  make all            Build and import the image"
