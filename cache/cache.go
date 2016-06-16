package cache

import (
	"time"

	"github.com/pivotal-golang/clock"
)

type item struct {
	value   string
	createdAt time.Time
}

type LocationCache struct {
	items map[string]*item
	clock clock.Clock
}

const CacheItemTTL = 60 * time.Second

func NewLocationCache(clock clock.Clock) *LocationCache {
	return &LocationCache{
		items: map[string]*item{},
		clock: clock,
	}
}

func (l *LocationCache) Lookup(repoName string) (string, bool) {
	if item, ok := l.items[repoName]; !ok {
		return "", false
	} else if l.clock.Since(item.createdAt) <= CacheItemTTL {
		return item.value, true
	}

	l.items[repoName] = nil

	return "", false
}

func (l *LocationCache) Add(repoName, location string) {
	l.items[repoName] = &item{value: location, createdAt: l.clock.Now()}
}
