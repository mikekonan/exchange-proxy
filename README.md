# freqtrade-proxy

Kucoin proxy for freqtrade that is using websockets to maintain candlestick/klines data in memory, thus having great
performance and reducing the amount of API calls to the Kucoin API. All other calls are proxied as usual.

This project I made just for myself but can add more exchanges in the future.

## OPS

### Local

```
git clone https://github.com/mikekonan/freqtrade-proxy.git
make build
./freqtrade-proxy -port 8080
```

#### config.json

```
{
    "exchange": {
        "name": "kucoin",
        "key": "",
        "secret": "",
        "ccxt_config": {
            "enableRateLimit": false,
            "urls": {
                "api": {
                    "public": "http://127.0.0.1:8080",
                    "private": "http://127.0.0.1:8080"
                }
            }
        },
        "ccxt_async_config": {
            "enableRateLimit": false
        }
    }
}
```

### Docker (suggested way)

```
docker run --restart=always -p 127.0.0.1:8080:8080 --name freqtrade-proxy -d mikekonan/freqtrade-proxy:main
```

#### config.json

```
{
    "exchange": {
        "name": "kucoin",
        "key": "",
        "secret": "",
        "ccxt_config": {
            "enableRateLimit": false,
            "urls": {
                "api": {
                    "public": "http://127.0.0.1:8080",
                    "private": "http://127.0.0.1:8080"
                }
            }
        },
        "ccxt_async_config": {
            "enableRateLimit": false
        }
    }
}
```

### Docker-compose (best way)

See example - [docker-compose.yml](docker-compose.yml)

```
  freqtrade-proxy:
    image: mikekonan/freqtrade-proxy:main
    restart: unless-stopped
    container_name: freqtrade-proxy
```

#### config.json

```
{
    "exchange": {
        "name": "kucoin",
        "key": "",
        "secret": "",
        "ccxt_config": {
            "enableRateLimit": false,
            "urls": {
                "api": {
                    "public": "http://freqtrade-proxy:8080/",
                    "private": "http://freqtrade-proxy:8080/"
                }
            }
        },
        "ccxt_async_config": {
            "enableRateLimit": false
        }
    }
}
```

## Donations

Donations are appreciated and will make me motivated to support and improve the project.

USDT TRC20 - TYssA3EUfAagJ9afF6vfwJvwwueTafMbGY

XRP - rNFugeoj3ZN8Wv6xhuLegUBBPXKCyWLRkB 1869777767

DOGE - D6xwe5V9jRkvWksiHiajwZsJ3KJxBVqBUC
