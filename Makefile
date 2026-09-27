COMPOSE_DIR := docker/local

up:
	cd $(COMPOSE_DIR) && docker compose up -d

down:
	cd $(COMPOSE_DIR) && docker compose down

logs:
	cd $(COMPOSE_DIR) && docker compose logs -f
