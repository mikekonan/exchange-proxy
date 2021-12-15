package kucoin

import (
	"fmt"
	"strings"
	"sync"
	"time"

	sdk "github.com/Kucoin/kucoin-go-sdk"
	"github.com/mikekonan/freqtradeProxy/proxy"
	"github.com/mikekonan/freqtradeProxy/store"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/valyala/fasthttp"
	"go.uber.org/ratelimit"
)

type subscriptionManager struct {
	clients []*ws
	rl      ratelimit.Limiter
	l       *sync.Mutex
	subs    map[string]struct{}
}

func (m *subscriptionManager) Subscribe(svc *sdk.ApiService, msg *sdk.WebSocketSubscribeMessage, store *store.Store, topicsPerWs int) {
	m.l.Lock()
	defer m.l.Unlock()

	if _, ok := m.subs[msg.Topic]; ok {
		return
	}

	m.subs[msg.Topic] = struct{}{}

	for i, c := range m.clients {
		if c.count == topicsPerWs {
			continue
		}

		c.count += 1
		m.rl.Take()
		if err := c.client.Subscribe(msg); err != nil {
			logrus.Fatal(err)
		}

		logrus.Infof("#%d-%d topic: '%s' subscribing...", i+1, c.count, msg.Topic)

		return
	}

	ws := newWs(svc, store)
	ws.count += 1
	if err := ws.client.Subscribe(msg); err != nil {
		logrus.Fatal(err)
	}

	m.clients = append(m.clients, ws)
	logrus.Infof("#%d-%d topic: '%s' subscribing...", len(m.clients), 1, msg.Topic)
}

func newWs(svc *sdk.ApiService, store *store.Store) *ws {
	resp, err := svc.WebSocketPublicToken()
	if err != nil {
		logrus.Fatal(err)
	}

	var token sdk.WebSocketTokenModel
	if err := resp.ReadData(&token); err != nil {
		logrus.Fatal(err)
	}

	wsClient := svc.NewWebSocketClientOpts(sdk.WebSocketClientOpts{Token: &token, Timeout: time.Second * 10, TLSSkipVerify: true})
	stream, errs, err := wsClient.Connect()
	if err != nil {
		logrus.Fatal(err)
	}

	result := &ws{client: wsClient, stream: stream, errs: errs}
	go result.serveFor(store)

	return result
}

type ws struct {
	client *sdk.WebSocketClient
	stream <-chan *sdk.WebSocketDownstreamMessage
	errs   <-chan error
	count  int
}

func New(store *store.Store, ttlCache *store.TTLCache, config *Config) *kucoin {
	client := &fasthttp.Client{
		ReadTimeout:  time.Second * 15,
		WriteTimeout: time.Second * 15,
	}

	//if config.Localaddr != "" {
	//	localAddr, err := net.ResolveIPAddr("ip", config.Localaddr)
	//	if err != nil {
	//		logrus.Errorf("cannot revolve '%v'", config.Localaddr)
	//	}
	//
	//	dialer := &net.Dialer{LocalAddr: &net.TCPAddr{IP: localAddr.IP}}
	//
	//	client.Dial = func(addr string) (net.Conn, error) {
	//		fmt.Println(addr)
	//		return dialer.Dial("tcp", addr)
	//	}
	//}

	instance := &kucoin{
		config:   config,
		client:   client,
		store:    store,
		ttlCache: ttlCache,
		subscriptionManager: &subscriptionManager{
			clients: nil,
			rl:      ratelimit.New(9),
			l:       new(sync.Mutex),
			subs:    map[string]struct{}{},
		},
	}

	svc := sdk.NewApiService(sdk.ApiKeyVersionOption(sdk.ApiKeyVersionV2))
	instance.svc = svc

	instance.rl = ratelimit.New(15)

	return instance
}

type kucoin struct {
	client *fasthttp.Client

	store    *store.Store
	ttlCache *store.TTLCache
	svc      *sdk.ApiService
	rl       ratelimit.Limiter

	subscriptionManager *subscriptionManager
	config              *Config
}

type candle struct {
	Symbol string         `json:"symbol"`
	Time   int64          `json:"time"`
	Candle sdk.KLineModel `json:"candles"`
}

func (ws *ws) serveFor(store *store.Store) {
	for {
		select {
		case err := <-ws.errs:
			logrus.Fatal("Error: %s", err.Error())
			return
		case msg := <-ws.stream:
			if msg == nil {
				continue
			}

			if strings.HasPrefix(msg.Topic, "/market/candles:") {
				candle := &candle{}
				err := msg.ReadData(candle)
				if err != nil {
					logrus.Fatal("cannot read candle data")
				}

				name := strings.Replace(msg.Topic, "/market/candles:", "", 1)
				pair := strings.Split(name, "_")[0]
				tf := strings.Split(name, "_")[1]

				store.Store(storeKey(pair, tf), timeframeToDuration(tf), parseCandle(candle.Candle))
			}
		}
	}
}

func (kucoin *kucoin) getKlines(pair string, timeframe string, startAt int64, endAt int64, retryCount int) (sdk.KLinesModel, *sdk.ApiResponse, error) {
	var (
		resp *sdk.ApiResponse
		err  error
	)

	for i := 1; i <= retryCount; i++ {
		kucoin.rl.Take()
		resp, err = kucoin.svc.KLines(pair, timeframe, startAt, endAt)
		if err == nil {
			break
		}

		if err != nil || i == retryCount || !strings.HasPrefix(resp.Code, "429") {
			return sdk.KLinesModel{}, resp, err
		}

		time.Sleep(time.Second)
	}

	candlesModel := sdk.KLinesModel{}
	if err := resp.ReadData(&candlesModel); err != nil {
		return candlesModel, resp, err
	}

	return candlesModel, resp, nil
}

func (kucoin *kucoin) transparentRequestURI(c *routing.Context) string {
	return fmt.Sprintf("%s/%s", kucoin.config.RequestURL, c.Request.URI().RequestURI()[8:])
}

func (kucoin *kucoin) Name() string {
	return "kucoin"
}

func (kucoin *kucoin) Routes() map[string]struct {
	Method  string
	Handler func(c *routing.Context) error
} {
	return map[string]struct {
		Method  string
		Handler func(c *routing.Context) error
	}{
		"api/v1/market/allTickers": {
			Method:  "GET",
			Handler: proxy.TransparentOverCacheHandler(kucoin.transparentRequestURI, kucoin.client, kucoin.ttlCache),
		},

		"api/v1/currencies": {
			Method:  "GET",
			Handler: proxy.TransparentOverCacheHandler(kucoin.transparentRequestURI, kucoin.client, kucoin.ttlCache),
		},

		"api/v1/symbols": {
			Method:  "GET",
			Handler: proxy.TransparentOverCacheHandler(kucoin.transparentRequestURI, kucoin.client, kucoin.ttlCache),
		},

		"*": {
			Method:  "GET",
			Handler: proxy.TransparentHandler(kucoin.transparentRequestURI, kucoin.client),
		},

		"api/v1/market/candles": {
			Method: "GET",
			Handler: func(c *routing.Context) error {
				logrus.Debugf("proxying - %s", c.Request.RequestURI())

				pair := string(c.Request.URI().QueryArgs().Peek("symbol"))
				timeframe := string(c.Request.URI().QueryArgs().Peek("type"))
				startAt := time.Unix(cast.ToInt64(string(c.Request.URI().QueryArgs().Peek("startAt"))), 0)
				endAt := time.Unix(cast.ToInt64(string(c.Request.URI().QueryArgs().Peek("endAt"))), 0)

				candles := kucoin.store.Get(storeKey(pair, timeframe), startAt, endAt)

				if len(candles) == 0 {
					candlesModel, resp, err := kucoin.getKlines(pair, timeframe, startAt.Unix(), endAt.Unix(), 15)
					if err != nil && resp == nil {
						logrus.Fatal(err)
					}

					c.Response.SetStatusCode(kucoinCodeToHttpCode(resp.Code))
					c.Response.SetBody((&apiResp{Code: resp.Code, RawData: resp.RawData, Message: resp.Message}).json())

					if len(candlesModel) == 0 {
						logrus.Warnf("there is no candle data from kucoin for - '%s'", c.Request.RequestURI())
					}

					kucoin.store.Store(
						storeKey(pair, timeframe),
						timeframeToDuration(timeframe),
						parseCandleModels(candlesModel)...,
					)

					if err == nil {
						go kucoin.subscriptionManager.Subscribe(
							kucoin.svc,
							sdk.NewSubscribeMessage(fmt.Sprintf("/market/candles:%s_%s", pair, timeframe), false),
							kucoin.store,
							kucoin.config.TopicsPerWs,
						)
					}

					return nil
				}

				_, err := c.Write(candles.KucoinRespJSON())
				return err
			},
		},
	}
}
