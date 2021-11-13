package kucoin

import (
	"fmt"
	"strings"
	"sync"
	"time"

	sdk "github.com/Kucoin/kucoin-go-sdk"
	"github.com/mikekonan/freqtradeProxy/model"
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
}

func (m *subscriptionManager) Subscribe(svc *sdk.ApiService, msg *sdk.WebSocketSubscribeMessage, store *store.Store) {
	m.l.Lock()
	defer m.l.Unlock()

	for i, c := range m.clients {
		if c.count == 100 {
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

func New(s *store.Store) *kucoin {
	instance := &kucoin{
		client:              fasthttp.Client{},
		store:               s,
		subscriptionManager: &subscriptionManager{clients: nil, rl: ratelimit.New(5), l: new(sync.Mutex)},
	}

	svc := sdk.NewApiService(sdk.ApiKeyVersionOption(sdk.ApiKeyVersionV2))
	instance.svc = svc

	instance.rl = ratelimit.New(10)

	return instance
}

type kucoin struct {
	client fasthttp.Client

	store *store.Store
	svc   *sdk.ApiService
	rl    ratelimit.Limiter

	subscriptionManager *subscriptionManager
}

func parseCandle(pair string, tf string, candle sdk.KLineModel) *model.Candle {
	return &model.Candle{
		Exchange:  "kucoin",
		Pair:      pair,
		Timeframe: tf,
		Ts:        cast.ToInt64(candle[0]),
		Open:      cast.ToFloat64(candle[1]),
		High:      cast.ToFloat64(candle[3]),
		Low:       cast.ToFloat64(candle[4]),
		Close:     cast.ToFloat64(candle[2]),
		Volume:    cast.ToFloat64(candle[5]),
		Amount:    cast.ToFloat64(candle[6]),
	}
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

				store.Store(parseCandle(pair, tf, candle.Candle))
			}
		}
	}
}

func (kucoin *kucoin) getKlines(pair string, timeframe string, startAt int64, endAt int64, retryCount int) (sdk.KLinesModel, error) {
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

		if i == retryCount {
			return sdk.KLinesModel{}, err
		}

		time.Sleep(time.Millisecond * 150)
	}

	candlesModel := sdk.KLinesModel{}
	if err := resp.ReadData(&candlesModel); err != nil {
		return candlesModel, err
	}

	return candlesModel, nil
}

func (kucoin *kucoin) timeframeToDuration(timeframe string) time.Duration {
	switch timeframe {
	case "1min":
		return time.Minute
	case "3min":
		return time.Minute * 3
	case "5min":
		return time.Minute * 5
	case "15min":
		return time.Minute * 15
	case "30min":
		return time.Minute * 30
	case "1hour":
		return time.Hour
	case "2hour":
		return time.Hour * 2
	case "4hour":
		return time.Hour * 4
	case "6hour":
		return time.Hour * 6
	case "8hour":
		return time.Hour * 8
	case "12hour":
		return time.Hour * 12
	case "1day":
		return time.Hour * 24
	}

	return time.Hour * 24 * 7
}

func (kucoin *kucoin) truncateTs(timeframe string, ts time.Time) time.Time {
	return ts.Truncate(kucoin.timeframeToDuration(timeframe))
}

func (kucoin *kucoin) Start(port int) {
	router := routing.New()

	router.Get("/kucoin/api/v1/market/candles", func(c *routing.Context) error {
		logrus.Infof("processing request - %s", c.Request.RequestURI())

		pair := string(c.Request.URI().QueryArgs().Peek("symbol"))
		timeframe := string(c.Request.URI().QueryArgs().Peek("type"))
		startAt := cast.ToInt64(string(c.Request.URI().QueryArgs().Peek("startAt")))
		endAt := cast.ToInt64(string(c.Request.URI().QueryArgs().Peek("endAt")))
		startTruncated := kucoin.truncateTs(timeframe, time.Unix(startAt, 0).UTC())
		endTruncated := kucoin.truncateTs(timeframe, time.Unix(endAt, 0).UTC())
		now := time.Now().UTC()
		if now.Before(time.Unix(endAt, 0).UTC()) {
			endTruncated = kucoin.truncateTs(timeframe, now)
		}

		candles := kucoin.store.Get("kucoin", pair, timeframe, startTruncated, endTruncated, kucoin.timeframeToDuration(timeframe))

		if len(candles) == 0 {
			candlesModel, err := kucoin.getKlines(pair, timeframe, startTruncated.Unix(), endAt, 3)
			if err != nil {
				return err
			}

			for _, c := range candlesModel {
				pc := parseCandle(pair, timeframe, *c)
				candles = append(candles, pc)
				kucoin.store.Store(pc)
			}

			kucoin.subscriptionManager.Subscribe(
				kucoin.svc,
				sdk.NewSubscribeMessage(fmt.Sprintf("/market/candles:%s_%s", pair, timeframe), false),
				kucoin.store,
			)
		}

		_, err := c.Write(candles.KucoinRespJSON())
		return err
	})

	router.Any("/kucoin/*", func(c *routing.Context) error {
		logrus.Infof("proxying over - %s", c.Request.RequestURI())

		req := fasthttp.AcquireRequest()
		c.Request.Header.CopyTo(&req.Header)
		req.SetRequestURI(fmt.Sprintf("https://openapi-v2.kucoin.com/%s", c.Request.URI().RequestURI()[8:]))
		req.SetBody(c.Request.Body())

		resp := fasthttp.AcquireResponse()
		if err := kucoin.client.Do(req, resp); err != nil {
			logrus.Error(err)
			return err
		}

		resp.Header.CopyTo(&c.Response.Header)
		c.Response.SetStatusCode(resp.StatusCode())
		c.Response.SetBody(resp.Body())

		return nil
	})

	logrus.Infof("starting proxy server on :%d port...", port)

	panic(fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), router.HandleRequest))
}
