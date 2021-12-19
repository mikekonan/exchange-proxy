# Freqtrade OPS

### Local
```shell
git clone https://github.com/mikekonan/freqtrade-proxy.git
make build
./freqtrade-proxy -port 8080 -verbose 1
```

#### config.json

```json
{
    "exchange": {
        "name": "kucoin",
        "key": "",
        "secret": "",
        "ccxt_config": {
            "enableRateLimit": false,
            "timeout": 60000,
            "urls": {
                "api": {
                    "public": "http://127.0.0.1:8080/kucoin",
                    "private": "http://127.0.0.1:8080/kucoin"
                }
            }
        },
        "ccxt_async_config": {
            "enableRateLimit": false,
            "timeout": 60000
        }
    }
}
```

### Docker (suggested way)

###### Use different tags for different platforms e.g. - main-amd64, main-arm-v6, main-arm-v7, main-arm64

```shell
docker run --restart=always -p 127.0.0.1:8080:8080 --name freqtrade-proxy -d mikekonan/freqtrade-proxy:main-amd64
```

#### config.json

```json
{
    "exchange": {
        "name": "kucoin",
        "key": "",
        "secret": "",
        "ccxt_config": {
            "enableRateLimit": false,
            "timeout": 60000,
            "urls": {
                "api": {
                    "public": "http://127.0.0.1:8080/kucoin",
                    "private": "http://127.0.0.1:8080/kucoin"
                }
            }
        },
        "ccxt_async_config": {
            "enableRateLimit": false,
            "timeout": 60000
        }
    }
}
```

### Docker-compose (best way)

###### Use different tags for different platforms e.g. - main-amd64, main-arm-v6, main-arm-v7, main-arm64

See example - [docker-compose.yml](freqtrade-docker-compose.yml)
```yaml
  freqtrade-proxy:
    image: mikekonan/freqtrade-proxy:main-amd64
    restart: unless-stopped
    container_name: freqtrade-proxy
    command: -verbose 1
```

#### config.json

```json
{
    "exchange": {
        "name": "kucoin",
        "key": "",
        "secret": "",
        "ccxt_config": {
            "enableRateLimit": false,
            "timeout": 60000,
            "urls": {
                "api": {
                    "public": "http://freqtrade-proxy:8080/kucoin",
                    "private": "http://freqtrade-proxy:8080/kucoin"
                }
            }
        },
        "ccxt_async_config": {
            "enableRateLimit": false,
            "timeout": 60000
        }
    }
}
```
