.DEFAULT_GOAL := list

.PHONY: list
list:
	@grep '^[^#[:space:]].*:' Makefile

.PHONY: setup
setup:
	go install github.com/air-verse/air@latest
	go mod download
	npm install

.PHONY: dev
dev:
	npx tailwindcss -i tailwind.css -o ./internal/dist/app.css
	npx concurrently -n frontend,go,browser -c blue,green,yellow \
		"tailwindcss -i tailwind.css -o ./internal/dist/app.css --watch" \
		"air -c .air.toml" \
		"wait-on http-get://localhost:8080 && open-cli http://localhost:8080"

.PHONY: build
build:
	npx tailwindcss -i tailwind.css -o ./internal/dist/app.css --minify
	CGO_ENABLED=0 go build -trimpath -tags timetzdata -o calendar ./cmd/calendar
