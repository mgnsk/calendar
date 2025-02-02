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
	cp ./node_modules/htmx.org/dist/htmx.min.js ./internal/dist/htmx.min.js
	npx concurrently -n frontend,go,browser -c blue,green,yellow \
		"tailwindcss -i tailwind.css -o ./internal/dist/app.css --watch" \
		"air -c .air.toml" \
		"wait-on http-get://localhost:8080 && open-cli http://localhost:8080"
