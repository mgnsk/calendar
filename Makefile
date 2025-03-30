.DEFAULT_GOAL := list

.PHONY: list
list:
	@grep '^[^#[:space:]].*:' Makefile

.PHONY: setup
setup:
	go install github.com/air-verse/air@latest
	go install github.com/mgechev/revive@latest
	go mod download
	npm ci

.PHONY: dev
dev:
	npx @tailwindcss/cli -i tailwind.css -o app.css
	npx concurrently -n tailwind,go,browser -c blue,green,yellow --kill-others-on-fail \
		"npx @tailwindcss/cli -i tailwind.css -o app.css --watch" \
		"air -c .air.toml" \
		"npx wait-on tcp:calendar.testing:8443 && npx open-cli https://calendar.testing:8443"


.PHONY: lint
lint:
	@npx concurrently --raw=true --group \
		"revive -max_open_files=64 -formatter=stylish -exclude=./vendor/... -config=revive.toml ./..." \
		"eslint"

.PHONY: build
build:
	docker build -t ghcr.io/mgnsk/calendar:edge .
