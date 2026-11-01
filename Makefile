COMPOSE_DIR := docker/local
TMUX_SESSION ?= wallet-services

.PHONY: up down logs logs-obs \
	build-eligibility run-eligibility test-eligibility \
	build-kyc run-kyc test-kyc \
	build-onboard run-onboard test-onboard \
	build-cust-client run-cust-client test-cust-client \
	build-customer run-customer test-customer \
	build-frontend run-frontend test-frontend \
	build-staff run-staff test-staff \
	start start-all start-backend stop stop-all restart attach status

up:
	cd $(COMPOSE_DIR) && docker compose up -d
	@echo "Observability Stack:"
	@echo "  - Grafana:     http://localhost:3002 (admin / admin)"
	@echo "  - Loki:        http://localhost:3100"
	@echo "  - Temporal UI: http://localhost:8080"
	@echo "  - Postgres:    localhost:5432"

down:
	cd $(COMPOSE_DIR) && docker compose down

logs:
	cd $(COMPOSE_DIR) && docker compose logs -f

logs-obs:
	cd $(COMPOSE_DIR) && docker compose logs -f loki grafana promtail

build-eligibility:
	cd ms-eligiblity && $(MAKE) build

run-eligibility:
	cd ms-eligiblity && $(MAKE) run

test-eligibility:
	cd ms-eligiblity && $(MAKE) test
build-kyc:
	cd ms-kyc && ./gradlew build

run-kyc:
	cd ms-kyc && ./gradlew bootRun

test-kyc:
	cd ms-kyc && ./gradlew test

build-onboard:
	cd ms-customer-onboard && $(MAKE) build

run-onboard:
	cd ms-customer-onboard && $(MAKE) run

test-onboard:
	cd ms-customer-onboard && $(MAKE) test

build-cust-client:
	cd ms-cust-onboard-client && $(MAKE) build

run-cust-client:
	cd ms-cust-onboard-client && $(MAKE) run

test-cust-client:
	cd ms-cust-onboard-client && $(MAKE) test

build-customer:
	cd cust-onboard-miniapp && pnpm build

run-customer:
	cd cust-onboard-miniapp && pnpm dev

test-customer:
	cd cust-onboard-miniapp && pnpm test

build-frontend: build-customer
run-frontend: run-customer
test-frontend: test-customer

build-staff:
	cd staff-manage-onboard-miniapp && pnpm build

run-staff:
	cd staff-manage-onboard-miniapp && pnpm dev

test-staff:
	cd staff-manage-onboard-miniapp && pnpm test

start:
	@command -v tmux >/dev/null 2>&1 || { echo "Error: tmux is not installed. Please install tmux first."; exit 1; }
	@if tmux has-session -t $(TMUX_SESSION) 2>/dev/null; then \
		echo "Tmux session '$(TMUX_SESSION)' is already running."; \
		echo "  make attach  -> view service logs"; \
		echo "  make stop    -> stop all services"; \
	else \
		echo "Starting all services in tmux session '$(TMUX_SESSION)'..."; \
		tmux new-session -d -s $(TMUX_SESSION) -n "eligibility" -c "$$(pwd)"; \
		tmux send-keys -t $(TMUX_SESSION):eligibility "make run-eligibility" C-m; \
		tmux new-window -t $(TMUX_SESSION) -n "kyc" -c "$$(pwd)"; \
		tmux send-keys -t $(TMUX_SESSION):kyc "make run-kyc" C-m; \
		tmux new-window -t $(TMUX_SESSION) -n "onboard" -c "$$(pwd)"; \
		tmux send-keys -t $(TMUX_SESSION):onboard "make run-onboard" C-m; \
		tmux new-window -t $(TMUX_SESSION) -n "cust-client" -c "$$(pwd)"; \
		tmux send-keys -t $(TMUX_SESSION):cust-client "make run-cust-client" C-m; \
		tmux new-window -t $(TMUX_SESSION) -n "customer-ui" -c "$$(pwd)"; \
		tmux send-keys -t $(TMUX_SESSION):customer-ui "make run-customer" C-m; \
		tmux new-window -t $(TMUX_SESSION) -n "staff-ui" -c "$$(pwd)"; \
		tmux send-keys -t $(TMUX_SESSION):staff-ui "make run-staff" C-m; \
		tmux select-window -t $(TMUX_SESSION):0; \
		echo "✓ All services started in tmux session '$(TMUX_SESSION)'."; \
		echo ""; \
		echo "Commands:"; \
		echo "  make attach       - View service logs (Ctrl+b then 0-5 to switch, Ctrl+b d to detach)"; \
		echo "  make status       - Check status of service windows"; \
		echo "  make stop         - Stop all services"; \
		echo "  make restart      - Restart all services"; \
	fi

start-all: start

start-backend:
	@command -v tmux >/dev/null 2>&1 || { echo "Error: tmux is not installed. Please install tmux first."; exit 1; }
	@if tmux has-session -t $(TMUX_SESSION) 2>/dev/null; then \
		echo "Tmux session '$(TMUX_SESSION)' is already running. Run 'make stop' first."; \
	else \
		echo "Starting backend services in tmux session '$(TMUX_SESSION)'..."; \
		tmux new-session -d -s $(TMUX_SESSION) -n "eligibility" -c "$$(pwd)"; \
		tmux send-keys -t $(TMUX_SESSION):eligibility "make run-eligibility" C-m; \
		tmux new-window -t $(TMUX_SESSION) -n "kyc" -c "$$(pwd)"; \
		tmux send-keys -t $(TMUX_SESSION):kyc "make run-kyc" C-m; \
		tmux new-window -t $(TMUX_SESSION) -n "onboard" -c "$$(pwd)"; \
		tmux send-keys -t $(TMUX_SESSION):onboard "make run-onboard" C-m; \
		tmux new-window -t $(TMUX_SESSION) -n "cust-client" -c "$$(pwd)"; \
		tmux send-keys -t $(TMUX_SESSION):cust-client "make run-cust-client" C-m; \
		tmux select-window -t $(TMUX_SESSION):0; \
		echo "✓ Backend services started in tmux session '$(TMUX_SESSION)'."; \
	fi

stop:
	@if tmux has-session -t $(TMUX_SESSION) 2>/dev/null; then \
		echo "Stopping tmux session '$(TMUX_SESSION)'..."; \
		tmux kill-session -t $(TMUX_SESSION); \
		echo "✓ All services stopped."; \
	else \
		echo "Tmux session '$(TMUX_SESSION)' is not running."; \
	fi

stop-all: stop

restart: stop
	@sleep 1
	@$(MAKE) start

attach:
	@if tmux has-session -t $(TMUX_SESSION) 2>/dev/null; then \
		tmux attach-session -t $(TMUX_SESSION); \
	else \
		echo "Tmux session '$(TMUX_SESSION)' is not running. Run 'make start' first."; \
	fi

status:
	@if tmux has-session -t $(TMUX_SESSION) 2>/dev/null; then \
		echo "Active windows in tmux session '$(TMUX_SESSION)':"; \
		tmux list-windows -t $(TMUX_SESSION) -F "  [#I] #W (active pane command: #{pane_current_command})"; \
	else \
		echo "Tmux session '$(TMUX_SESSION)' is not running."; \
	fi



