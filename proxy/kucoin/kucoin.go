package kucoin

import (
	"fmt"
	"strings"
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

func New(s *store.Store) *kucoin {
	return &kucoin{
		client: fasthttp.Client{},
		store:  s,
	}
}

type kucoin struct {
	client fasthttp.Client

	ws     *sdk.WebSocketClient
	stream <-chan *sdk.WebSocketDownstreamMessage
	errs   <-chan error

	store *store.Store
	svc   *sdk.ApiService
	rl    ratelimit.Limiter
	wsRl  ratelimit.Limiter

	subCount int
}

func (kucoin *kucoin) Connect() error {
	svc := sdk.NewApiService(sdk.ApiKeyVersionOption(sdk.ApiKeyVersionV2))
	kucoin.svc = svc

	kucoin.rl = ratelimit.New(20)
	kucoin.wsRl = ratelimit.New(9)

	resp, err := kucoin.svc.WebSocketPublicToken()
	if err != nil {
		return err
	}

	var token sdk.WebSocketTokenModel
	if err := resp.ReadData(&token); err != nil {
		return err
	}

	kucoin.ws = kucoin.svc.NewWebSocketClientOpts(sdk.WebSocketClientOpts{Token: &token, Timeout: time.Minute})

	stream, errs, err := kucoin.ws.Connect()
	if err != nil {
		return err
	}

	kucoin.stream = stream
	kucoin.errs = errs

	return nil
}

func (kucoin *kucoin) parseCandle(pair string, tf string, candle sdk.KLineModel) *model.Candle {
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

func (kucoin *kucoin) Start() {
	router := routing.New()

	go func() {
		for {
			select {
			case err := <-kucoin.errs:
				kucoin.ws.Stop()
				logrus.Fatal("Error: %s", err.Error())
				return
			case msg := <-kucoin.stream:
				if strings.HasPrefix(msg.Topic, "/market/candles:") {
					candle := &candle{}
					err := msg.ReadData(candle)
					if err != nil {
						logrus.Fatal("cannot read candle data")
					}

					name := strings.Replace(msg.Topic, "/market/candles:", "", 1)
					pair := strings.Split(name, "_")[0]
					tf := strings.Split(name, "_")[1]

					kucoin.store.Store(kucoin.parseCandle(pair, tf, candle.Candle))

					return
				}
			}
		}
	}()

	router.Get("/api/v1/market/candles", func(c *routing.Context) error {
		pair := string(c.Request.URI().QueryArgs().Peek("symbol"))
		timeframe := string(c.Request.URI().QueryArgs().Peek("type"))
		startAt := cast.ToInt64(string(c.Request.URI().QueryArgs().Peek("startAt")))
		endAt := cast.ToInt64(string(c.Request.URI().QueryArgs().Peek("endAt")))

		logrus.Infof("%s-%s-%d-%d", pair, timeframe, startAt, endAt)

		candles := kucoin.store.Get("kucoin", pair, timeframe, startAt, endAt)

		if len(candles) == 0 {
			kucoin.rl.Take()
			resp, err := kucoin.svc.KLines(pair, timeframe, startAt, endAt)
			if err != nil {
				return err
			}

			candlesModel := sdk.KLinesModel{}
			if err := resp.ReadData(&candlesModel); err != nil {
				logrus.Fatal(err)
			}

			for _, c := range candlesModel {
				pc := kucoin.parseCandle(pair, timeframe, *c)
				candles = append(candles, pc)
				kucoin.store.Store(pc)
			}

			kucoin.wsRl.Take()
			err = kucoin.ws.Subscribe(
				sdk.NewSubscribeMessage(fmt.Sprintf("/market/candles:%s_%s", pair, timeframe), false),
			)

			kucoin.subCount++
			logrus.Warn(kucoin.subCount)

			if err != nil {
				return err
			}
		}

		c.Write(candles.KucoinRespJSON())

		return nil
	})

	router.Get("*", func(c *routing.Context) error {
		req := fasthttp.AcquireRequest()
		c.Request.Header.CopyTo(&req.Header)

		req.SetRequestURI(fmt.Sprintf("https://openapi-v2.kucoin.com/%s", c.Request.URI().RequestURI()))
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

	panic(fasthttp.ListenAndServe(":8080", router.HandleRequest))
}
