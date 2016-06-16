package cache_test

import (
	"time"

	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/pivotal-golang/clock/fakeclock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lookup", func() {
	var locationCache *cache.LocationCache
	var clock *fakeclock.FakeClock

	BeforeEach(func() {
		clock = fakeclock.NewFakeClock(time.Now())
		locationCache = cache.NewLocationCache(clock)
	})

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
		})
	})
})
