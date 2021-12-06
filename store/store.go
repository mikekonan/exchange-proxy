package store

import (
	"database/sql"
	_ "embed"
	"errors"
	"sync"
	"time"

	"go4.org/sort"

	"github.com/jmoiron/sqlx"
	"github.com/mikekonan/freqtradeProxy/model"
	"github.com/sirupsen/logrus"

	"github.com/doug-martin/goqu/v9"
	_ "modernc.org/sqlite"
)

const cacheSize = 5000

var g = goqu.Dialect("sqlite3")

//go:embed bootstrap.sql
var bootstrapScript string

func New() *Store {
	store := new(Store)
	var err error
	store.conn, err = sqlx.Open("sqlite", "file::memory:?cache=shared")
	//store.conn, err = sqlx.Open("sqlite", "kek.db")
	if err != nil {
		logrus.Panic(err)
	}

	if _, err := store.conn.Exec(bootstrapScript); err != nil {
		logrus.Panic(err)
	}

	store.m = new(sync.Mutex)

	return store
}

type Store struct {
	conn *sqlx.DB
	m    *sync.Mutex
}

func (store *Store) selectCandleCountTsQuery(exchange string, timeframe string, pair string, ts int64) string {
	query, _, _ := g.Select(goqu.COUNT("*")).
		From("candles").
		Where(goqu.Ex{"exchange": exchange, "timeframe": timeframe, "pair": pair, "ts": ts}).
		ToSQL()

	return query
}

func (store *Store) selectCandleCountQuery(exchange string, timeframe string, pair string) string {
	query, _, _ := g.Select(goqu.COUNT("*")).
		From("candles").
		Where(goqu.Ex{"exchange": exchange, "timeframe": timeframe, "pair": pair}).
		ToSQL()

	return query
}

func (store *Store) updateByRowIdQuery(close float64, ts int64, open float64, high float64, low float64, volume float64, amout float64, rowid int) string {
	query, _, _ := g.Update("candles").
		Set(goqu.Record{
			"close":  close,
			"ts":     ts,
			"open":   open,
			"high":   high,
			"low":    low,
			"volume": volume,
			"amount": amout,
		}).
		Where(goqu.Ex{"rowid": rowid}).
		ToSQL()

	return query
}

func (store *Store) selectRowIdQuery(exchange string, timeframe string, pair string) string {
	query, _, _ := g.Select("rowid", goqu.MIN("ts")).
		From("candles").
		Where(goqu.Ex{"exchange": exchange, "timeframe": timeframe, "pair": pair}).Limit(1).
		ToSQL()

	return query
}

func (stroe *Store) updateCandleQuery(open float64, high float64, low float64, close float64, volume float64, amount float64, exchange string, timeframe string, pair string, ts int64) string {
	query, _, _ := g.Update("candles").
		Set(goqu.Record{"open": open, "high": high, "low": low, "close": close, "volume": volume, "amount": amount}).
		Where(goqu.Ex{"exchange": exchange, "timeframe": timeframe, "pair": pair, "ts": ts}).
		ToSQL()

	return query
}

func (stroe *Store) insertCandlesQuery(exchange string, pair string, timeframe string, ts int64, open float64, high float64, low float64, close float64, volume float64, amount float64) string {
	query, _, _ := g.Insert("candles").
		Cols("exchange", "pair", "timeframe", "ts", "open", "high", "low", "close", "volume", "amount").
		Vals(goqu.Vals{exchange, pair, timeframe, ts, open, high, low, close, volume, amount}).
		ToSQL()

	return query
}

func (store *Store) Store(candle model.Candle) {
	store.m.Lock()
	defer store.m.Unlock()

	logrus.Tracef("storing candle for %s at %s of %s-%s", candle.Exchange, time.Unix(candle.Ts, 0), candle.Pair, candle.Timeframe)
	tx, err := store.conn.Beginx()
	if err != nil {
		logrus.Panic(err)
	}

	defer tx.Commit()

	result := tx.QueryRow(store.selectCandleCountTsQuery(candle.Exchange, candle.Timeframe, candle.Pair, candle.Ts))
	if result.Err() != nil {
		logrus.Panic(result.Err())
	}

	var count int
	if err := result.Scan(&count); err != nil {
		logrus.Panic(err)
	}

	if count > 0 {
		if _, err := tx.Exec(store.updateCandleQuery(candle.Open, candle.High, candle.Low, candle.Close, candle.Volume, candle.Amount, candle.Exchange, candle.Timeframe, candle.Pair, candle.Ts)); err != nil {
			logrus.Panic(err)
		}

		return
	}

	result = tx.QueryRow(store.selectCandleCountQuery(candle.Exchange, candle.Timeframe, candle.Pair))
	if result.Err() != nil {
		logrus.Panic(result.Err())
	}

	if err := result.Scan(&count); err != nil {
		logrus.Panic(err)
	}

	if count == cacheSize {
		result = tx.QueryRow(store.selectRowIdQuery(candle.Exchange, candle.Timeframe, candle.Pair))
		if result.Err() != nil {
			logrus.Panic(result.Err())
		}

		var (
			rowid int
			minTs int
		)

		if err := result.Scan(&rowid, &minTs); err != nil {
			logrus.Panic(err)
		}

		if _, err := tx.Exec(store.updateByRowIdQuery(candle.Close, candle.Ts, candle.Open, candle.High, candle.Low, candle.Volume, candle.Amount, rowid)); err != nil {
			logrus.Panic(err)
		}

		return
	}

	if _, err := tx.Exec(store.insertCandlesQuery(candle.Exchange, candle.Pair, candle.Timeframe, candle.Ts, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume, candle.Amount)); err != nil {
		logrus.Panic(err)
	}
}

func (store *Store) selectCandlesQuery(exchange string, pair string, timeframe string, from int64, to int64) string {
	query, _, _ := g.Select("*").From("candles").
		Where(
			goqu.Ex{"exchange": exchange, "timeframe": timeframe, "pair": pair},
			goqu.C("ts").Between(goqu.Range(from, to)),
		).
		Order(goqu.I("ts").Asc()).
		ToSQL()

	return query
}

func (store *Store) selectCandleQuery(exchange string, pair string, timeframe string, from int64) string {
	query, _, _ := g.Select("*").From("candles").
		Where(
			goqu.Ex{"exchange": exchange, "timeframe": timeframe, "pair": pair},
			goqu.C("ts").Eq(from),
		).
		ToSQL()

	return query
}

func (store *Store) selectLastCandleBeforeTs(exchange string, pair string, timeframe string, ts int64) string {
	query, _, _ := g.Select("*").From("candles").
		Where(
			goqu.Ex{"exchange": exchange, "timeframe": timeframe, "pair": pair},
			goqu.C("ts").Lt(ts),
		).
		Order(goqu.I("ts").Desc()).Limit(1).
		ToSQL()

	return query
}

func (store *Store) selectFirstCandleAfterTs(exchange string, pair string, timeframe string, ts int64) string {
	query, _, _ := g.Select("*").From("candles").
		Where(
			goqu.Ex{"exchange": exchange, "timeframe": timeframe, "pair": pair},
			goqu.C("ts").Gt(ts),
		).
		Order(goqu.I("ts").Asc()).Limit(1).
		ToSQL()

	return query
}

func (store *Store) Get(exchange string, pair string, timeframe string, from time.Time, to time.Time, period time.Duration) (result model.Candles) {
	store.m.Lock()
	defer store.m.Unlock()

	tx, err := store.conn.Beginx()
	if err != nil {
		logrus.Panic(err)
	}

	defer tx.Commit()

	fromCandle := model.Candle{}
	if err := tx.Get(&fromCandle, store.selectCandleQuery(exchange, pair, timeframe, from.Unix())); err != nil && !errors.Is(err, sql.ErrNoRows) {
		logrus.Panic()
	}

	if fromCandle.Ts == 0 {
		if err := tx.Get(&fromCandle, store.selectLastCandleBeforeTs(exchange, pair, timeframe, from.Unix())); err != nil && !errors.Is(err, sql.ErrNoRows) {
			logrus.Panic()
		}

		if fromCandle.Ts == 0 {
			return nil
		}

		fromCandle.Ts = from.Unix()
	}

	var tsCandles = make(map[int64]model.Candle, cacheSize)
	tsCandles[fromCandle.Ts] = fromCandle

	var storedCandles []model.Candle
	if err := tx.Select(&storedCandles, store.selectCandlesQuery(exchange, pair, timeframe, from.Unix(), to.Unix())); err != nil {
		logrus.Panic(err)
	}

	for _, r := range storedCandles {
		tsCandles[r.Ts] = r
	}

	prevCandle := fromCandle
	for i := fromCandle.Ts; i <= to.Unix(); i += int64(period.Seconds()) {
		if _, ok := tsCandles[i]; !ok {
			candle := prevCandle
			candle.Ts = i
			candle.Volume = 0
			candle.Amount = 0
			tsCandles[i] = candle
		} else {
			prevCandle = tsCandles[i]
		}
	}

	result = make(model.Candles, 0, cacheSize)
	for k := range tsCandles {
		result = append(result, tsCandles[k])
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Ts > result[j].Ts
	})

	return result
}
