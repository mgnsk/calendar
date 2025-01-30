.DEFAULT_GOAL := list

.PHONY: list
list:
	@grep '^[^#[:space:]].*:' Makefile

.PHONY: setup
setup:
	go install github.com/air-verse/air@latest
	go mod download
	npm ci

.PHONY: dev
dev:
	npx tailwindcss -i tailwind.css -o ./internal/dist/app.css
	npx concurrently -n frontend,go,browser -c blue,green,yellow \
		"tailwindcss -i tailwind.css -o ./internal/dist/app.css --watch" \
		"air -c .air.toml" \
		"wait-on http-get://localhost:8080 && open-cli http://localhost:8080"
