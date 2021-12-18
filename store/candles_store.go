package store

import (
	"sync"
	"time"

	"github.com/mikekonan/freqtradeProxy/model"
	"github.com/sirupsen/logrus"
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
		if bucket.first != nil {
			steps := c.Ts.Sub(bucket.first.value.Ts) / period

			if steps > 1 {
				for i := 1; i < int(steps); i++ {
					painted := bucket.first.value.Clone()
					painted.Ts = painted.Ts.Add(period)
					painted.Volume = 0
					painted.Amount = 0

					logrus.Warnf("saving painted candle: ts '%s' for '%s'...", painted.Ts, key)

					s.store(bucket, painted)
				}
			}
		}

		s.store(bucket, c)
	}
}

func (s *Store) store(bucket *candlesLinkedList, candle *model.Candle) {
	first, ok := bucket.get(0)
	if ok && first.Ts == candle.Ts {
		logrus.Debugf("%s %s - update first", first.Ts.String(), candle.Ts.String())
		bucket.set(0, candle)

		return
	}

	if bucket.size() == s.cacheSize {
		bucket.remove(s.cacheSize - 1)
	}

	if ok && first.Ts.Before(candle.Ts) {
		logrus.Debugf("%s %s - prepend", first.Ts.String(), candle.Ts.String())
		bucket.prepend(candle)
	} else {
		if first != nil {
			logrus.Debugf("%s %s - append", first.Ts.String(), candle.Ts.String())
		}

		bucket.append(candle)
	}
}

func (s *Store) Get(key string, from time.Time, to time.Time) []*model.Candle {
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
