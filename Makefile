# Único comando del repo. En el host solo hace falta Docker y Make.
COMPOSE ?= docker compose
HOST_IP ?= $(shell ip route get 1.1.1.1 2>/dev/null | awk '{for (i=1;i<=NF;i++) if ($$i=="src") {print $$(i+1); exit}}')

.DEFAULT_GOAL := help

.PHONY: help doctor up down stop logs ps restart \
	db api web mobile test test-api test-client psql \
	build clean desktop

help: ## lista los targets
	@awk 'BEGIN {FS = ":.*##"; printf "\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  make %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf "\nEn el host alcanza con Docker y Make.\n"
	@printf "Web http://localhost:3000   API http://localhost:8080\n\n"

doctor: ## chequea que Docker y Compose estén
	@command -v docker >/dev/null || { echo "falta docker"; exit 1; }
	@$(COMPOSE) version >/dev/null
	@echo "ok: docker + compose"

up: ## db + api + web (todo en contenedores)
	$(COMPOSE) up --build -d db api web
	@echo
	@echo "web  http://localhost:3000"
	@echo "api  http://localhost:8080/api/hello"

down: ## para y borra contenedores
	$(COMPOSE) --profile mobile --profile test down

stop: ## para sin borrar volúmenes
	$(COMPOSE) --profile mobile --profile test stop

logs: ## logs de db/api/web
	$(COMPOSE) logs -f db api web

ps: ## contenedores
	$(COMPOSE) ps

restart: ## rebuild y levanta de nuevo
	$(COMPOSE) up --build -d db api web

db: ## solo Postgres
	$(COMPOSE) up -d db

api: ## solo API (+ db)
	$(COMPOSE) up --build -d db api

web: ## solo web (+ api + db)
	$(COMPOSE) up --build -d db api web

mobile: ## Metro/Expo en Docker. En el teléfono: Expo Go
	HOST_IP="$(HOST_IP)" EXPO_PUBLIC_API_URL="http://$(HOST_IP):8080" \
		$(COMPOSE) --profile mobile up --build mobile

test: test-api test-client ## todos los tests en contenedores

test-api: ## go test del API
	$(COMPOSE) --profile test run --rm api-test

test-client: ## tests del cliente de sync
	$(COMPOSE) --profile test run --rm --build client-test

psql: ## shell SQL
	$(COMPOSE) exec db psql -U hola -d hola

build: ## construye imágenes sin levantar
	$(COMPOSE) --profile mobile --profile test build

desktop: ## Electron no se dockeriza (necesita UI nativa)
	@echo "Electron (Win/Lin/Mac) no corre bien dentro de Docker."
	@echo "La UI es la misma que web: make up && abrir http://localhost:3000"
	@echo "Si igual querés el binario nativo, hace falta Node en el host:"
	@echo "  pnpm install && pnpm --filter @hola/desktop dev"

clean: ## borra contenedores, imágenes del proyecto y volúmenes
	$(COMPOSE) --profile mobile --profile test down -v --rmi local
