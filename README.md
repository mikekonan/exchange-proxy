# exchange-proxy

Exchange proxy using WebSockets to maintain candlestick/klines data in memory, thus having great performance, reducing the number of API calls to the exchange API, decreases latency and CPU usage.
There is no warranty of correct working. You take all risks of using this.
All improvements are made by me on a voluntary basis in my spare time.

## OPS

### Usage
```shell
Usage of ./dist/exchange-proxy:
  -bindaddr string
        bindable address (default "0.0.0.0")
  -cache-size int
        amount of candles to cache (default 1000)
  -client-timeout duration
        client timeout (default 15s)
  -concurrency-limit int
        server concurrency limit (default 262144)
  -kucoin-api-url string
        kucoin api address (default "https://openapi-v2.kucoin.com")
  -kucoin-topics-per-ws int
        amount of topics per ws connection [10-280] (default 200)
  -port string
        listen port (default "8080")
  -ttl-cache-timeout duration
        ttl of blobs of cached data (default 10m0s)
  -verbose int
        verbose level: 0 - info, 1 - debug, 2 - trace
```

#### Note
All unforeseen connection errors or the inaccessibility of the exchange will lead to the proxy crash, which means that you have to handle it on your end 

### Local
```shell
./exchange-proxy -port 8080
```

### Docker (suggested way)

###### Use different tags for different platforms e.g. - latest-amd64, latest-arm-v6, latest-arm-v7, latest-arm64

```shell
docker run --restart=always -p 127.0.0.1:8080:8080 --name exchange-proxy -d mikekonan/exchange-proxy:latest-amd64
```

#### Examples of usage:
- [freqtrade](./docs/ops/freqtrade.md)

# Supported exchanges:
- [Kucoin](./docs/exchanges/kucoin.md)

## Donations

Donations are appreciated and will make me motivated to support and improve the project.

- USDT TRC20 - TYssA3EUfAagJ9afF6vfwJvwwueTafMbGY
- XRP - rNFugeoj3ZN8Wv6xhuLegUBBPXKCyWLRkB 1869777767
- DOGE - D6xwe5V9jRkvWksiHiajwZsJ3KJxBVqBUC
- BTC - 35SrQDWAfwXcRGHaKbxNWwvHRNSLAbVjrk
- ETH - 0x37c34bac13cf60f022be1bdea2dec1136cdc838a


### Referral links:
- [Kucoin](https://www.kucoin.com/ucenter/signup?rcode=rJ327D3)

- [Okex](https://www.okex.com/join/3941527)

- [Gate.io](https://www.gate.io/signup/3325373)

- [Currency.com](https://currency.com/trading/signup?c=ciqjuj5y&pid=referral)
