package kucoin

import (
	"fmt"
	"time"

	sdk "github.com/Kucoin/kucoin-go-sdk"
	"github.com/mikekonan/freqtradeProxy/model"
	"github.com/spf13/cast"
)

func timeframeToDuration(timeframe string) time.Duration {
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

func storeKey(pair string, tf string) string {
	return fmt.Sprintf("kucoin-%s-%s", pair, tf)
}

func kucoinCodeToHttpCode(str string) int {
	if len(str) < 3 {
		return 200
	}

	return cast.ToInt(str[:3])
}

func parseCandle(candle sdk.KLineModel) *model.Candle {
	return &model.Candle{
		Ts:     time.Unix(cast.ToInt64(candle[0]), 0).UTC(),
		Open:   cast.ToFloat64(candle[1]),
		High:   cast.ToFloat64(candle[3]),
		Low:    cast.ToFloat64(candle[4]),
		Close:  cast.ToFloat64(candle[2]),
		Volume: cast.ToFloat64(candle[5]),
		Amount: cast.ToFloat64(candle[6]),
	}
}

func parseCandleModels(candlesModel sdk.KLinesModel) model.Candles {
	candles := make(model.Candles, 0, len(candlesModel))
	for _, c := range candlesModel {
		pc := parseCandle(*c)
		candles = append(candles, pc)
	}

	return candles
}
