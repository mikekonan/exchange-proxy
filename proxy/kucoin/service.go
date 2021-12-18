package kucoin

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mailru/easyjson"
	"github.com/mikekonan/freqtradeProxy/proxy"
	"github.com/mikekonan/freqtradeProxy/store"
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

func New(store *store.Store, ttlCache *store.TTLCache, client *proxy.Client, config *Config) *service {
	httpRl := ratelimit.New(15)

	instance := &service{
		config:   config,
		client:   client,
		store:    store,
		ttlCache: ttlCache,
		rl:       httpRl,
		subscriber: &subscriber{
			l:      new(sync.Mutex),
			pool:   nil,
			httpRl: httpRl,
			wsRl:   ratelimit.New(9),
			subs:   map[string]struct{}{},
			config: config,
			client: client,
			store:  store,
		},
	}

	return instance
}

type service struct {
	client *proxy.Client

	store    *store.Store
	ttlCache *store.TTLCache
	rl       ratelimit.Limiter

	subscriber *subscriber
	config     *Config
}

func (service *service) executeKLinesRequest(pair string, timeframe string, startAt int64, endAt int64) (int, *kLinesResponse, []byte, error) {
	path := fmt.Sprintf("%s/%s?type=%s&symbol=%s&startAt=%d&endAt=%d", service.config.RequestURL, kLinesPath, timeframe, pair, startAt, endAt)

	statusCode, data, err := service.client.Get(nil, path)
	if err != nil {
		return statusCode, nil, nil, err
	}

	kLinesResponse := &kLinesResponse{}
	if err := easyjson.Unmarshal(data, kLinesResponse); err != nil {
		return statusCode, nil, data, err
	}

	return statusCode, kLinesResponse, data, nil
}

func (service *service) getKlines(pair string, timeframe string, startAt int64, endAt int64, retryCount int) (int, *kLinesResponse, []byte, error) {
	for i := 1; i <= retryCount; i++ {
		service.rl.Take()

		if statusCode, kLinesResponse, data, err := service.executeKLinesRequest(pair, timeframe, startAt, endAt); statusCode == 200 {
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

func (service *service) transparentRequestURI(c *routing.Context) string {
	return fmt.Sprintf("%s/%s", service.config.RequestURL, c.Request.URI().RequestURI()[8:])
}

func (service *service) Name() string {
	return "kucoin"
}

func (service *service) Routes() []struct {
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
			Method:  http.MethodGet,
			Handler: proxy.TransparentOverCacheHandler(service.transparentRequestURI, service.client, service.ttlCache),
		},

		{
			Path:    currenciesPath,
			Method:  http.MethodGet,
			Handler: proxy.TransparentOverCacheHandler(service.transparentRequestURI, service.client, service.ttlCache),
		},

		{
			Path:    symbolsPath,
			Method:  http.MethodGet,
			Handler: proxy.TransparentOverCacheHandler(service.transparentRequestURI, service.client, service.ttlCache),
		},

		{
			Path:   kLinesPath,
			Method: http.MethodGet,
			Handler: func(c *routing.Context) error {
				logrus.Debugf("proxying - %s", c.Request.RequestURI())

				pair := string(c.Request.URI().QueryArgs().Peek("symbol"))
				timeframe := string(c.Request.URI().QueryArgs().Peek("type"))
				startAt := time.Unix(cast.ToInt64(string(c.Request.URI().QueryArgs().Peek("startAt"))), 0)
				endAt := time.Unix(cast.ToInt64(string(c.Request.URI().QueryArgs().Peek("endAt"))), 0)
				endAtAfterNow := endAt.After(time.Now().UTC())

				candles := service.store.Get(storeKey(pair, timeframe), startAt, endAt)

				if len(candles) == 0 {
					statusCode, klinesResponse, data, err := service.getKlines(pair, timeframe, startAt.Unix(), endAt.Unix(), 15)

					c.Response.SetStatusCode(statusCode)
					c.Response.SetBody(data)

					if statusCode == 429 {
						return nil
					}

					if len(klinesResponse.Klines) == 0 {
						logrus.Warnf("there is no candle data from kucoin for - '%s'", c.Request.RequestURI())
					}

					if endAtAfterNow {
						service.store.Store(
							storeKey(pair, timeframe),
							timeframeToDuration(timeframe),
							parseKLines(klinesResponse.Klines)...,
						)

						if err == nil {
							go service.subscriber.subscribeKLines(pair, timeframe)
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
			Handler: proxy.TransparentHandler(service.transparentRequestURI, service.client),
		},
	}
}
