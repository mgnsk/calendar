.DEFAULT_GOAL := list

UID := $(shell id -u)
GID := $(shell id -g)

# List all commands in this file.
.PHONY: list
list:
	@grep '^[^#[:space:]].*:' Makefile

# Set up the development environment.
.PHONY: setup
setup:
	# Build the sandbox.
	docker compose -f docker-compose.dev.yml build \
		--build-arg uid=${UID} \
		--build-arg gid=${GID}

	# Install tools and dependencies.
	docker compose -f docker-compose.dev.yml run --rm sandbox \
		go mod download
	docker compose -f docker-compose.dev.yml run --rm sandbox \
		go install tool
	docker compose -f docker-compose.dev.yml run --rm sandbox \
		npm ci

# Start the development environment.
.PHONY: dev
dev:
	docker compose -f docker-compose.dev.yml up -d

	# Build the CSS once.
	docker compose -f docker-compose.dev.yml exec sandbox \
		npx @tailwindcss/cli -i tailwind.css -o app.css

	# Start the application and automatic rebuild.
	docker compose -f docker-compose.dev.yml exec sandbox \
		npx concurrently -n tailwind,go -c blue,green,yellow --kill-others-on-fail \
		"npx @tailwindcss/cli -i tailwind.css -o app.css --watch" \
		"air -c .air.toml"

# Stop the development environment.
.PHONY: stop
stop:
	docker compose -f docker-compose.dev.yml down

# Lint application code.
.PHONY: lint
lint:
	docker compose -f docker-compose.dev.yml run --rm sandbox \
		npx concurrently --raw=true --group \
		"revive -formatter=stylish -exclude=./vendor/... -config=revive.toml ./..." \
		"eslint"

# Build the production docker image.
.PHONY: build
build:
	docker build -t ghcr.io/mgnsk/calendar:edge .

# Install and enable the production service.
.PHONY: install
install:
	docker volume create "calendar-database" || true
	ln -s ./mgnsk-calendar.service /etc/systemd/system/mgnsk-calendar.service
	systemctl enable --now mgnsk-calendar.service
