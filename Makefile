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
	npx tailwindcss -i tailwind.css -o app.css
	npx concurrently -n tailwind,go,browser -c blue,green,yellow \
		"tailwindcss -i tailwind.css -o app.css --watch" \
		"air -c .air.toml" \
		"wait-on tcp:calendar.testing:8443 && open-cli https://calendar.testing:8443"


.PHONY: lint
lint:
	@npx concurrently --raw=true --group \
		"revive -max_open_files=64 -formatter=stylish -exclude=./vendor/... -config=revive.toml ./..." \
		"eslint"
