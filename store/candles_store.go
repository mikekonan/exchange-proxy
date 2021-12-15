package store

import (
	"sync"
	"time"

	"github.com/mikekonan/freqtradeProxy/model"
)

func NewStore(cacheSize int) *Store {
	return &Store{
		l:           new(sync.RWMutex),
		mappedLists: map[string]*candlesLinkedList{},
		cacheSize:   cacheSize,
	}
}

type Store struct {
	l           *sync.RWMutex
	mappedLists map[string]*candlesLinkedList
	cacheSize   int
}

func (s *Store) Store(key string, period time.Duration, candles ...*model.Candle) {
	s.l.Lock()
	defer s.l.Unlock()

	bucket := s.mappedLists[key]
	if bucket == nil {
		bucket = newCandlesLinkedList()
		s.mappedLists[key] = bucket
	}

	for _, c := range candles {
		if bucket.last != nil {
			steps := c.Ts.Sub(bucket.last.value.Ts) / period

			if steps > 1 {
				for i := 1; i < int(steps); i++ {
					painted := bucket.last.value.Clone()
					painted.Ts = painted.Ts.Add(time.Duration(i) * period)
					painted.Volume = 0
					painted.Amount = 0
					s.store(key, painted)
				}
			}
		}

		s.store(key, c)
	}
}

func (s *Store) store(key string, candle *model.Candle) {
	bucket := s.mappedLists[key]

	first, ok := bucket.get(0)
	if ok && first.Ts == candle.Ts {
		bucket.set(0, candle)

		return
	}

	if bucket.size() == s.cacheSize {
		bucket.remove(s.cacheSize - 1)
	}

	bucket.prepend(candle)

	return
}

func (s *Store) Get(key string, from time.Time, to time.Time) model.Candles {
	s.l.RLock()
	defer s.l.RUnlock()

	bucket := s.mappedLists[key]
	if bucket == nil {
		return nil
	}

	candles := bucket.selectInRangeReversedFn(
		func(candle *model.Candle) bool { return candle.Ts == from || candle.Ts.Before(from) },
		func(candle *model.Candle) bool { return candle.Ts == to || candle.Ts.Before(to) },
	)

	return candles
}
