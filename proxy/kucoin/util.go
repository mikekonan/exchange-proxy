package kucoin

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/mikekonan/exchange-proxy/model"
	"github.com/spf13/cast"
)

var (
	startArrayJsonBytes = []byte(`[`)
	endArrayJsonBytes   = []byte(`]`)
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

func parseCandle(candle kLine) *model.Candle {
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

func parseKLines(candlesModel kLines) []*model.Candle {
	candles := make([]*model.Candle, 0, len(candlesModel))

	for _, c := range candlesModel {
		pc := parseCandle(*c)
		candles = append(candles, pc)
	}

	return candles
}

func floatFmt(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func candlesJSON(candles []*model.Candle) []byte {
	buff := bytes.NewBuffer(nil)
	buff.Write(startArrayJsonBytes)
	for _, c := range candles {
		buff.Write([]byte(fmt.Sprintf(`["%d","%s","%s","%s","%s","%s","%s"],`, c.Ts.Unix(), floatFmt(c.Open), floatFmt(c.Close), floatFmt(c.High), floatFmt(c.Low), floatFmt(c.Volume), floatFmt(c.Amount))))
	}

	if len(candles) > 0 {
		buff.Truncate(buff.Len() - 1)
	}

	buff.Write(endArrayJsonBytes)

	return buff.Bytes()
}

func wsTopic(pair string, tf string) string {
	return fmt.Sprintf("%s_%s", pair, tf)
}
