# Freqtrade OPS

### Local

```shell
git clone https://github.com/mikekonan/exchange-proxy.git
make build
./exchange-proxy -port 8080 -verbose 1
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

###### Use different tags for different platforms e.g. - latest-amd64, latest-arm-v6, latest-arm-v7, latest-arm64

```shell
docker run --restart=always -p 127.0.0.1:8080:8080 --name exchange-proxy -d mikekonan/exchange-proxy:latest-amd64
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

###### Use different tags for different platforms e.g. - latest-amd64, latest-arm-v6, latest-arm-v7, latest-arm64

See example - [docker-compose.yml](freqtrade-docker-compose.yml)

```yaml
  exchange-proxy:
    image: mikekonan/exchange-proxy:latest-amd64
    restart: unless-stopped
    container_name: exchange-proxy
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
          "public": "http://exchange-proxy:8080/kucoin",
          "private": "http://exchange-proxy:8080/kucoin"
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
