.PHONY: help build clean

.DEFAULT_GOAL := help

build: ## build binaries
	go build -trimpath -o ./dist/freqtrade-proxy

clean: ## clean
	rm ./dist/freqtrade-proxy*

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
