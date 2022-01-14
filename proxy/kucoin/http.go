package kucoin

import (
	"fmt"
	netHttp "net/http"
	"sync"
	"time"

	"github.com/mailru/easyjson"
	"github.com/mikekonan/exchange-proxy/proxy"
	"github.com/mikekonan/exchange-proxy/store"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"go.uber.org/ratelimit"
)

const (
	kLinesPath     = "api/v1/market/candles"
	tickersPath    = "api/v1/market/allTickers"
	currenciesPath = "api/v1/currencies"
	symbolsPath    = "api/v1/symbols"
)

func New(store *store.Store, ttlCache *store.TTLCache, client *proxy.Client, config *Config) *http {
	httpRl := ratelimit.New(15)

	instance := &http{
		config:   config,
		client:   client,
		store:    store,
		ttlCache: ttlCache,
		rl:       httpRl,
		subscriber: &subscriber{
			l:      new(sync.Mutex),
			pool:   nil,
			httpRl: httpRl,
			wsRl:   ratelimit.New(5),
			subs:   map[string]struct{}{},
			config: config,
			client: client,
			store:  store,
		},
	}

	return instance
}

type http struct {
	client *proxy.Client

	store    *store.Store
	ttlCache *store.TTLCache
	rl       ratelimit.Limiter

	subscriber *subscriber
	config     *Config
}

func (http *http) executeKLinesRequest(pair string, timeframe string, startAt int64, endAt int64) (int, *kLinesResponse, []byte, error) {
	path := fmt.Sprintf("%s/%s?type=%s&symbol=%s&startAt=%d&endAt=%d", http.config.KucoinApiURL, kLinesPath, timeframe, pair, startAt, endAt)

	statusCode, data, err := http.client.Get(nil, path)
	if err != nil {
		return statusCode, nil, nil, err
	}

	kLinesResponse := &kLinesResponse{}
	if err := easyjson.Unmarshal(data, kLinesResponse); err != nil {
		return statusCode, nil, data, err
	}

	return statusCode, kLinesResponse, data, nil
}

func (http *http) getKlines(pair string, timeframe string, startAt int64, endAt int64, retryCount int) (int, *kLinesResponse, []byte, error) {
	for i := 1; i <= retryCount; i++ {
		http.rl.Take()

		if statusCode, kLinesResponse, data, err := http.executeKLinesRequest(pair, timeframe, startAt, endAt); statusCode == 200 {
			return statusCode, kLinesResponse, data, nil
		} else {
			if i == retryCount {
				return statusCode, kLinesResponse, data, fmt.Errorf("get klines request '%s' '%s' '%d' '%d' exceeded retry '%d' attemts: %w", pair, timeframe, startAt, endAt, retryCount, err)
			}

			time.Sleep(time.Second)
		}
	}

	return 500, nil, nil, fmt.Errorf("retry count is zero")
}

func (http *http) transparentRequestURI(c *routing.Context) string {
	return fmt.Sprintf("%s/%s", http.config.KucoinApiURL, c.Request.URI().RequestURI()[8:])
}

func (http *http) Name() string {
	return "kucoin"
}

func (http *http) Routes() []struct {
	Path    string
	Method  string
	Handler func(c *routing.Context) error
} {

	return []struct {
		Path    string
		Method  string
		Handler func(c *routing.Context) error
	}{
		{
			Path:    tickersPath,
			Method:  netHttp.MethodGet,
			Handler: proxy.TransparentOverCacheHandler(http.transparentRequestURI, http.client, http.ttlCache),
		},

		{
			Path:    currenciesPath,
			Method:  netHttp.MethodGet,
			Handler: proxy.TransparentOverCacheHandler(http.transparentRequestURI, http.client, http.ttlCache),
		},

		{
			Path:    symbolsPath,
			Method:  netHttp.MethodGet,
			Handler: proxy.TransparentOverCacheHandler(http.transparentRequestURI, http.client, http.ttlCache),
		},

		{
			Path:   kLinesPath,
			Method: netHttp.MethodGet,
			Handler: func(c *routing.Context) error {
				logrus.Debugf("proxying - %s", c.Request.RequestURI())

				pair := string(c.Request.URI().QueryArgs().Peek("symbol"))
				timeframe := string(c.Request.URI().QueryArgs().Peek("type"))
				startAt := time.Unix(cast.ToInt64(string(c.Request.URI().QueryArgs().Peek("startAt"))), 0)
				endAt := time.Unix(cast.ToInt64(string(c.Request.URI().QueryArgs().Peek("endAt"))), 0)
				endAtAfterNow := endAt.After(time.Now().UTC())

				candles := http.store.Get(storeKey(pair, timeframe), startAt, endAt)

				if len(candles) == 0 {
					statusCode, klinesResponse, data, err := http.getKlines(pair, timeframe, startAt.Unix(), endAt.Unix(), 15)

					c.Response.SetStatusCode(statusCode)
					c.Response.SetBody(data)

					if statusCode == 429 {
						return nil
					}

					if len(klinesResponse.Klines) == 0 {
						logrus.Warnf("there is no candle data from kucoin for - '%s'", c.Request.RequestURI())
					}

					if endAtAfterNow {
						http.store.Store(
							storeKey(pair, timeframe),
							timeframeToDuration(timeframe),
							parseKLines(klinesResponse.Klines)...,
						)

						if err == nil {
							go http.subscriber.subscribeKLines(pair, timeframe)
						}
					}

					return nil
				}

				data, err := easyjson.Marshal(genericResponse{Code: "200000", Data: candlesJSON(candles)})

				if err != nil {
					return err
				}

				c.SetStatusCode(200)
				c.SetBody(data)

				return err
			},
		},

		{
			Path:    "*",
			Method:  proxy.AnyHTTPMethod,
			Handler: proxy.TransparentHandler(http.transparentRequestURI, http.client),
		},
	}
}
