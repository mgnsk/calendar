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
	cp ./node_modules/mark.js/dist/mark.min.js ./internal/dist/mark.min.js
	npx concurrently -n tailwind,go,browser -c blue,green,yellow \
		"tailwindcss -i tailwind.css -o ./internal/dist/app.css --watch" \
		"air -c .air.toml" \
		"wait-on tcp:calendar.testing:8443 && open-cli https://calendar.testing:8443"
