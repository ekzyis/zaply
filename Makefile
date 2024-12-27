.PHONY: dev

dev:
	bash livereload.sh

build:
	templ generate
	go build -o zaply main.go
