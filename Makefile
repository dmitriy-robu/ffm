APP_NAME=api
WS_NAME=ws
ROOT_PATH=/app
DOCKER_EXE=docker compose
DOCKER_NETWORK=fitness-network

.PHONY: docker build check-network create-env

docker: check-network create-env
	@echo "Starting the Docker Compose stack..."
	$(DOCKER_EXE) up -d

build:
	@echo "Building API server..."
	@cd "$(ROOT_PATH)/cmd/$(APP_NAME)" && go build -buildvcs=false -o "$(ROOT_PATH)/tmp"
	@chmod +x "$(ROOT_PATH)/tmp"

check-network:
	@echo "Checking if the network exists..."
	@docker network inspect $(DOCKER_NETWORK) > /dev/null 2>&1 || (echo "Network does not exist. Creating..." && docker network create $(DOCKER_NETWORK))

create-env:
	@if [ ! -f .env ]; then echo "Creating .env file from example..."; cp .env.example .env; fi
