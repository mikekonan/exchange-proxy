.PHONY: help lint build clean clean-generated generate

.DEFAULT_GOAL := help

generate: clean-generated ## generate
	go generate ./...
	go fmt ./...

build: generate ## build binaries
	go build -trimpath -o ./dist/freqtrade-proxy

clean-generated: ## clean generated
	find . -name '*_easyjson.go' -delete

clean: ## clean
	rm -rf ./dist/freqtrade-proxy*
	find . -name '*_easyjson.go' -delete

lint: ## lint
	golangci-lint run --path-prefix $(PWD)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
