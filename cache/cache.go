package cache

import (
	"time"

	"github.com/pivotal-golang/clock"
)

type cacheEntry struct {
	// location is the full url to the repo on github
	location  string
	updatedAt time.Time
}

type LocationCache struct {
	items map[string]*cacheEntry
	clock clock.Clock
}

const CacheItemTTL = 15 * time.Minute

func NewLocationCache(clock clock.Clock) *LocationCache {
	return &LocationCache{
		items: map[string]*cacheEntry{},
		clock: clock,
	}
}

func (l *LocationCache) Lookup(repoName string) (string, bool) {
	if item, ok := l.items[repoName]; !ok {
		return "", false
	} else if l.clock.Since(item.updatedAt) <= CacheItemTTL {
		return item.location, true
	}

	delete(l.items, repoName)

	return "", false
}

func (l *LocationCache) Add(repoName, location string) {
	l.items[repoName] = &cacheEntry{location: location, updatedAt: l.clock.Now()}
}
