.PHONY: help build build-linux clean

.DEFAULT_GOAL := help

build: ## build binary
	go build -trimpath -o ./bin/freqtrade-proxy

clean: ## clean
	rm ./bin/freqtrade-proxy

build-linux: ## build linux-binary
	docker run -v $(PWD):/tmp/src golang:buster /bin/bash -c "cd /tmp/src && make build"

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
