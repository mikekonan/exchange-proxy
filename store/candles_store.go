package store

import (
	"sort"
	"sync"
	"time"

	"github.com/mikekonan/freqtradeProxy/model"
)

func NewCandlesStore(cacheSize int) *CandlesStore {
	return &CandlesStore{
		l:           new(sync.RWMutex),
		mappedLists: map[string]*candlesLinkedList{},
		cacheSize:   cacheSize,
	}
}

type CandlesStore struct {
	l           *sync.RWMutex
	mappedLists map[string]*candlesLinkedList
	cacheSize   int
}

func (s *CandlesStore) Store(key string, candles ...*model.Candle) {
	s.l.Lock()
	defer s.l.Unlock()

	if s.mappedLists[key] == nil {
		s.mappedLists[key] = newCandlesLinkedList()
	}

	for _, c := range candles {
		s.store(key, c)
	}
}

func (s *CandlesStore) store(key string, candle *model.Candle) {
	first, ok := s.mappedLists[key].get(0)
	if ok && first.Ts == candle.Ts {
		s.mappedLists[key].set(0, candle)

		return
	}

	if s.mappedLists[key].size() == s.cacheSize {
		s.mappedLists[key].remove(s.cacheSize - 1)
	}

	s.mappedLists[key].prepend(candle)

	return
}

func (s *CandlesStore) Get(key string, from time.Time, to time.Time, period time.Duration) model.Candles {
	s.l.RLock()
	defer s.l.RUnlock()

	bucket := s.mappedLists[key]
	if bucket == nil {
		return nil
	}

	candles := bucket.selectInRangeReversedFn(
		func(candle *model.Candle) bool {
			return candle.Ts == from || candle.Ts.Before(from)
		},
		func(candle *model.Candle) bool {
			return candle.Ts == to
		},
	)

	if len(candles) == 0 {
		return nil
	}

	firstCandle := candles[len(candles)-1]

	var tsCandles = make(map[time.Time]*model.Candle, s.cacheSize)
	tsCandles[firstCandle.Ts] = firstCandle

	for _, r := range candles {
		tsCandles[r.Ts] = r
	}

	prevCandle := firstCandle
	for i := firstCandle.Ts; i.Before(to) || i == to; i = i.Add(period) {
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

	result := make(model.Candles, 0, s.cacheSize)
	for k := range tsCandles {
		result = append(result, tsCandles[k])
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Ts.After(result[j].Ts)
	})

	return candles
}
