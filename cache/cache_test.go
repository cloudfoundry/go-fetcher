package cache_test

import (
	"time"

	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Location Cache", func() {
	var locationCache *cache.LocationCache
	var clock *fakeclock.FakeClock

	BeforeEach(func() {
		clock = fakeclock.NewFakeClock(time.Now())
		logger := lagertest.NewTestLogger("cache")
		locationCache = cache.NewLocationCache(logger, clock)
	})

	Describe("Lookup", func() {
		Context("when there is nothing in the cache", func() {
			It("returns not ok", func() {
				_, ok := locationCache.Lookup("something")
				Expect(ok).To(BeFalse())
			})
		})

		Context("when looking up a value that is in the cache", func() {
			BeforeEach(func() {
				locationCache.Add("repo-name", "cached-location")
			})

			It("returns ok", func() {
				_, ok := locationCache.Lookup("repo-name")
				Expect(ok).To(BeTrue())
			})

			It("returns the cached location", func() {
				location, _ := locationCache.Lookup("repo-name")
				Expect(location).To(Equal("cached-location"))
			})

			It("returns not ok after TTL expires", func() {
				clock.Increment(cache.CacheItemTTL + 1)
				_, ok := locationCache.Lookup("repo-name")
				Expect(ok).To(BeFalse())
				_, ok = locationCache.Lookup("repo-name")
				Expect(ok).To(BeFalse())
			})
		})
	})

	Describe("Add", func() {
		Context("when there is already an entry for the given key", func() {
			BeforeEach(func() {
				locationCache.Add("repo-name", "cached-location")
			})

			It("updates the TTL", func() {
				clock.Increment(cache.CacheItemTTL - 1)
				locationCache.Add("repo-name", "cached-location")

				clock.Increment(cache.CacheItemTTL)
				location, _ := locationCache.Lookup("repo-name")
				Expect(location).To(Equal("cached-location"))
			})
		})
	})
})
