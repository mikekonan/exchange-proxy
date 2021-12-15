package store

import (
	"sync"
	"time"
)

type Container struct {
	raw       []byte
	isGzipped bool
	expiresAt time.Time
}

func (c *Container) Raw() []byte {
	return c.raw
}

func (c *Container) IsGzipped() bool {
	return c.isGzipped
}

func NewTTLCache(expirationTimeout time.Duration) *TTLCache {
	return &TTLCache{
		l:                 new(sync.Mutex),
		kv:                map[string]*Container{},
		expirationTimeout: expirationTimeout,
	}
}

type TTLCache struct {
	l *sync.Mutex

	kv                map[string]*Container
	expirationTimeout time.Duration
}

func (s *TTLCache) Get(key string) *Container {
	s.l.Lock()
	defer s.l.Unlock()

	container, ok := s.kv[key]
	if !ok {
		return nil
	}

	if container.expiresAt.Before(time.Now().UTC()) {
		s.kv[key] = nil
		return nil
	}

	return container
}

func (s *TTLCache) Store(key string, value []byte, isGzipped bool) {
	s.l.Lock()
	defer s.l.Unlock()

	s.kv[key] = &Container{
		raw:       value,
		expiresAt: time.Now().UTC().Add(s.expirationTimeout),
		isGzipped: isGzipped,
	}
}
