package store

import (
	"database/sql"
	_ "embed"
	"sync"

	"github.com/mikekonan/freqtradeProxy/model"
	"github.com/sirupsen/logrus"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/mattn/go-sqlite3"
)

const cacheSize = 1500

var g = goqu.Dialect("sqlite3")

//go:embed bootstrap.sql
var bootstrapScript string

func New() *Store {
	store := new(Store)
	var err error
	store.conn, err = sql.Open("sqlite3", "file::memory:?cache=shared")
	//store.conn, err = sql.Open("sqlite3", "kek.db")
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
	conn *sql.DB
	m    *sync.Mutex
}

func (store *Store) Store(candle *model.Candle) {
	store.m.Lock()
	defer store.m.Unlock()

	tx, err := store.conn.Begin()
	if err != nil {
		logrus.Panic(err)
	}

	defer tx.Commit()

	query, _, _ := g.Select(goqu.COUNT("*")).
		From("candles").
		Where(goqu.Ex{"exchange": candle.Exchange, "timeframe": candle.Timeframe, "pair": candle.Pair, "ts": candle.Ts}).
		ToSQL()

	result := tx.QueryRow(query)
	if result.Err() != nil {
		logrus.Panic(result.Err())
	}

	var count int
	if err := result.Scan(&count); err != nil {
		logrus.Panic(err)
	}

	if count > 0 {
		query, _, _ := g.Update("candles").
			Set(goqu.Record{"close": candle.Close, "volume": candle.Volume, "amount": candle.Amount}).
			Where(goqu.Ex{"exchange": candle.Exchange, "timeframe": candle.Timeframe, "pair": candle.Pair, "ts": candle.Ts}).
			ToSQL()

		if _, err := tx.Exec(query); err != nil {
			logrus.Panic(err)
		}

		return
	}

	query, _, _ = g.Select(goqu.COUNT("*")).
		From("candles").
		Where(goqu.Ex{"exchange": candle.Exchange, "timeframe": candle.Timeframe, "pair": candle.Pair}).
		ToSQL()

	result = tx.QueryRow(query)
	if result.Err() != nil {
		logrus.Panic(result.Err())
	}

	if err := result.Scan(&count); err != nil {
		logrus.Panic(err)
	}

	if count == cacheSize {
		query, _, _ = g.Select("rowid", goqu.MIN("ts")).
			From("candles").
			Where(goqu.Ex{"exchange": candle.Exchange, "timeframe": candle.Timeframe, "pair": candle.Pair}).Limit(1).
			ToSQL()

		result = tx.QueryRow(query)
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

		query, _, _ := g.Update("candles").
			Set(goqu.Record{
				"close":  candle.Close,
				"ts":     candle.Ts,
				"open":   candle.Open,
				"high":   candle.High,
				"low":    candle.Low,
				"volume": candle.Volume,
				"amount": candle.Amount,
			}).
			Where(goqu.Ex{"rowid": rowid}).
			ToSQL()

		if _, err := tx.Exec(query); err != nil {
			logrus.Panic(err)
		}

		return
	}

	query, _, _ = g.Insert("candles").
		Cols("exchange", "pair", "timeframe", "ts", "open", "high", "low", "close", "volume", "amount").
		Vals(goqu.Vals{candle.Exchange, candle.Pair, candle.Timeframe, candle.Ts, candle.Open, candle.High, candle.Low, candle.Close, candle.Volume, candle.Amount}).
		ToSQL()

	if _, err := tx.Exec(query); err != nil {
		logrus.Panic(err)
	}
}

func (store *Store) Get(exchange string, pair string, timeframe string, from int64, to int64) (result model.Candles) {
	store.m.Lock()
	defer store.m.Unlock()

	query, _, _ := g.Select("*").From("candles").
		Where(
			goqu.Ex{"exchange": exchange, "timeframe": timeframe, "pair": pair},
			goqu.C("ts").Between(goqu.Range(from, to)),
		).
		ToSQL()

	tx, err := store.conn.Begin()
	if err != nil {
		logrus.Panic(err)
	}

	defer tx.Commit()

	rows, err := tx.Query(query)
	if err != nil {
		logrus.Panic(err)
	}

	result = make([]*model.Candle, 0, cacheSize)

	for rows.Next() {
		var current model.Candle
		err = rows.Scan(
			&current.Exchange, &current.Pair, &current.Timeframe, &current.Ts,
			&current.Open, &current.High, &current.Low, &current.Close, &current.Volume, &current.Amount,
		)

		if err != nil {
			logrus.Panic(err)
		}

		result = append(result, &current)
	}

	return result
}
