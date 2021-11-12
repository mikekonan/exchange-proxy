# freqtrade-proxy
Kucoin proxy for freqtrade that is using websockets to maintain candlestick/klines data in memory, thus having great performance and reducing the amount of API calls to the Kucoin API. All other calls are proxied as usual.
This project I made just for myself but can add more exchanges in the future.

## USAGE
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

## Execution
```docker run --restart=always -p 169.254.1.1:8080:8080 --name freqtrade-proxy -d freqtrade-proxy```

## Donations
Donations are appreciated and will make me motivated to support and improve the project.

USDT TRC20 - TYssA3EUfAagJ9afF6vfwJvwwueTafMbGY

XRP - rNFugeoj3ZN8Wv6xhuLegUBBPXKCyWLRkB 1869777767

DOGE - D6xwe5V9jRkvWksiHiajwZsJ3KJxBVqBUC
