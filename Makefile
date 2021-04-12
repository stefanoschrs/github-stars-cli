#!make

upx := $(shell which upx)

run:
	STORAGE_FILE=~/.cache/github-stars-cli go run . list --username ${u}

build:
	go build -o github-stars-cli -ldflags="-w -s" .
ifdef upx
	upx github-stars-cli
endif

.PHONY: run build
