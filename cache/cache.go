package cache

import (
	"time"

	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/lager"
)

type cacheEntry struct {
	// location is the full url to the repo on github
	location  string
	updatedAt time.Time
}

type LocationCache struct {
	items  map[string]*cacheEntry
	logger lager.Logger
	clock  clock.Clock
}

func NewLocationCache(logger lager.Logger, clock clock.Clock) *LocationCache {
	return &LocationCache{
		items:  map[string]*cacheEntry{},
		logger: logger,
		clock:  clock,
	}
}

func (l *LocationCache) Lookup(repoName string) (string, bool) {
	if item, ok := l.items[repoName]; !ok {
		return "", false
	} else {
		return item.location, true
	}
}

func (l *LocationCache) Add(repoName, location string) {
	l.items[repoName] = &cacheEntry{location: location, updatedAt: l.clock.Now()}
}

func (l *LocationCache) Swap(newLocationCache *LocationCache) {
	logger := l.logger

	logger.Info("cache-items-swap", lager.Data{"old_len": len(l.items), "new_len": len(newLocationCache.items)})
	l.items = newLocationCache.items
}
