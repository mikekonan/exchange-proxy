.PHONY: help build build-local clean

.DEFAULT_GOAL := help

build-local:
	go build -trimpath -o ./bin/freqtrade-proxy

build: ## build binaries
	go build -trimpath -o ./bin/freqtrade-proxy-darwin-amd64
	for arch in "linux/amd64" "linux/arm/v6" "linux/arm/v7" "linux/arm64" ; do \
		echo $${arch//\//-}; \
		docker run --platform $$arch -v $(PWD):/tmp/src golang:buster /bin/bash -c "cd /tmp/src && go build -trimpath -o ./bin/freqtrade-proxy-"$${arch//\//-} ;\
	done

clean: ## clean
	rm ./bin/freqtrade-proxy*

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
