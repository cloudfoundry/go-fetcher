package cache_test

import (
	"time"

	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/cloudfoundry/go-fetcher/cache/fakes"
	"github.com/google/go-github/github"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CacheLoader", func() {
	var (
		fakeRepoService *fakes.FakeRepositoriesService
		cacheLoader     ifrit.Runner
		locCache        *cache.LocationCache
		fakeClock       *fakeclock.FakeClock
	)

	BeforeEach(func() {
		fakeRepoService = &fakes.FakeRepositoriesService{}
		fakeClock = fakeclock.NewFakeClock(time.Now())
		locCache = cache.NewLocationCache(clock.NewClock())
		cacheLoader = cache.NewCacheLoader([]string{"org1", "org2"}, locCache, fakeRepoService, fakeClock)
		fakeRepoService.ListByOrgReturns(nil, &github.Response{}, nil)
	})

	It("queries github before becoming ready", func() {
		cacheLoaderProcess := ifrit.Background(cacheLoader)

		doneCh := make(chan struct{})

		fakeRepoService.ListByOrgStub = func(org string, opt *github.RepositoryListByOrgOptions) ([]github.Repository, *github.Response, error) {
			<-doneCh
			return nil, &github.Response{}, nil
		}

		Consistently(cacheLoaderProcess.Ready()).ShouldNot(BeClosed())
		doneCh <- struct{}{}
		Consistently(cacheLoaderProcess.Ready()).ShouldNot(BeClosed())
		doneCh <- struct{}{}

		Eventually(cacheLoaderProcess.Ready()).Should(BeClosed())

		Expect(fakeRepoService.ListByOrgCallCount()).To(Equal(2))
	})

	It("requests all repos for each org", func() {
		ifrit.Invoke(cacheLoader)

		org, _ := fakeRepoService.ListByOrgArgsForCall(0)
		Expect(org).To(Equal("org1"))

		org, _ = fakeRepoService.ListByOrgArgsForCall(1)
		Expect(org).To(Equal("org2"))
	})

	It("stores the repos in the cache", func() {
		returnedRepos := []github.Repository{}

		name := "repo1"
		returnedRepos = append(returnedRepos, github.Repository{
			Name: &name,
		})
		name2 := "repo2"
		returnedRepos = append(returnedRepos, github.Repository{
			Name: &name2,
		})
		fakeRepoService.ListByOrgReturns(returnedRepos, &github.Response{}, nil)

		cacheLoader = cache.NewCacheLoader([]string{"http://github.com/org1/"}, locCache, fakeRepoService, fakeClock)
		ifrit.Invoke(cacheLoader)
		storedLocation, foundInCache := locCache.Lookup("repo1")

		Expect(foundInCache).To(BeTrue())
		Expect(storedLocation).To(Equal("http://github.com/org1/repo1"))
		storedLocation, foundInCache = locCache.Lookup("repo2")

		Expect(foundInCache).To(BeTrue())
		Expect(storedLocation).To(Equal("http://github.com/org1/repo2"))
	})

	It("follows the NextPage link in paginated results", func() {
		nextPage := 0
		fakeRepoService.ListByOrgStub = func(org string, opt *github.RepositoryListByOrgOptions) ([]github.Repository, *github.Response, error) {
			if nextPage == 3 {
				nextPage = 0 // 0 signals no more pages
			} else {
				nextPage++
			}
			return []github.Repository{
					github.Repository{},
					github.Repository{},
				},
				&github.Response{
					NextPage: nextPage,
					LastPage: 3,
				}, nil
		}

		ifrit.Invoke(cacheLoader)

		// Expect 8 calls because we have 2 organizations, and each have 4 pages (0,1,2,3)
		Expect(fakeRepoService.ListByOrgCallCount()).To(Equal(8))
	})

	It("updates the cache periodically", func() {
		ifrit.Invoke(cacheLoader)

		Expect(fakeRepoService.ListByOrgCallCount()).To(Equal(2))

		fakeClock.WaitForWatcherAndIncrement(cache.CacheUpdateInterval)
		Eventually(fakeRepoService.ListByOrgCallCount).Should(Equal(4))

		fakeClock.WaitForWatcherAndIncrement(cache.CacheUpdateInterval)
		Eventually(fakeRepoService.ListByOrgCallCount).Should(Equal(6))
	})
})
