# Kucoin

API docs:

- [Kucoin API docs](https://docs.kucoin.com)

## Proxy paths:

| Path                      | Methods | Comment                               |
|---------------------------|---------|---------------------------------------|
| /api/v1/market/candles    | GET     | cached in application store in memory |
| /api/v1/market/allTickers | GET     | cached as blob in memory              |
| /api/v1/currencies        | GET     | cached as blob in memory              |
| /api/v1/symbols           | GET     | cached as blob in memory              |
| *                         | ANY     | proxied transparently                 |

## Configuration

| Param                | Comment                                                                    |
|----------------------|----------------------------------------------------------------------------|
| kucoin-api-url       | kucoin api base URL                                                        |
| kucoin-topics-per-ws | amount of topics per ws connection. **recommended value between 100-250 ** |
| cache-size           | number of candles in application memory per {pair_tf}                      |
| ttl-cache-timeout    | cache blobs ttl                                                            |
