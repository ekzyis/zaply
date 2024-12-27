.PHONY: dev

dev:
	bash livereload.sh

build:
	templ generate
	tailwindcss -i ./public/css/input.css -o ./public/css/tailwind.css
	go build -o zaply main.go
