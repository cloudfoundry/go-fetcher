package cache_test

import (
	"errors"
	"time"

	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/cloudfoundry/go-fetcher/cache/fakes"
	"github.com/google/go-github/github"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/pivotal-golang/lager/lagertest"
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
		cacheLogger := lagertest.NewTestLogger("cache")
		locCache = cache.NewLocationCache(cacheLogger, clock.NewClock())
		logger := lagertest.NewTestLogger("cache-loader")
		cacheLoader = cache.NewCacheLoader(logger, "http://example.com", []string{"org1", "org2"}, locCache, fakeRepoService, fakeClock)
		fakeRepoService.ListByOrgReturns(nil, &github.Response{}, nil)
	})

	It("queries github before becoming ready", func() {
		doneCh := make(chan struct{})

		fakeRepoService.ListByOrgStub = func(org string, opt *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error) {
			<-doneCh
			return nil, &github.Response{}, nil
		}

		cacheLoaderProcess := ifrit.Background(cacheLoader)

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
		Expect(org).To(Equal("org2"))

		org, _ = fakeRepoService.ListByOrgArgsForCall(1)
		Expect(org).To(Equal("org1"))
	})

	It("stores the repos in the cache", func() {
		fakeRepoService.ListByOrgStub = func(org string, _ *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error) {
			if org == "org1" {
				name := "repo1"
				url := "http://example.com/org1/repo1"
				return []*github.Repository{{Name: &name, HTMLURL: &url}}, &github.Response{}, nil
			}

			if org == "org2" {
				name := "repo2"
				url := "http://example.com/org2/repo2"
				return []*github.Repository{{Name: &name, HTMLURL: &url}}, &github.Response{}, nil
			}
			return nil, nil, errors.New("not found")
		}

		ifrit.Invoke(cacheLoader)
		storedLocation, foundInCache := locCache.Lookup("repo1")

		Expect(foundInCache).To(BeTrue())
		Expect(storedLocation).To(Equal("http://example.com/org1/repo1"))
		storedLocation, foundInCache = locCache.Lookup("repo2")

		Expect(foundInCache).To(BeTrue())
		Expect(storedLocation).To(Equal("http://example.com/org2/repo2"))
	})

	It("follows the NextPage link in paginated results", func() {
		nextPage := 0
		fakeRepoService.ListByOrgStub = func(org string, opt *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error) {
			if nextPage == 3 {
				nextPage = 0 // 0 signals no more pages
			} else {
				nextPage++
			}
			return []*github.Repository{
					&github.Repository{},
					&github.Repository{},
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

	It("Prefers the first org for each repo", func() {
		fakeRepoService.ListByOrgStub = func(org string, _ *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error) {
			if org == "org1" {
				name := "repo1"
				url := "http://example.com/org1/repo1"
				return []*github.Repository{{Name: &name, HTMLURL: &url}}, &github.Response{}, nil
			}

			if org == "org2" {
				name := "repo1"
				url := "http://example.com/org2/repo1"
				return []*github.Repository{{Name: &name, HTMLURL: &url}}, &github.Response{}, nil
			}
			return nil, nil, errors.New("not found")
		}

		ifrit.Invoke(cacheLoader)
		storedLocation, foundInCache := locCache.Lookup("repo1")

		Expect(foundInCache).To(BeTrue())
		Expect(storedLocation).To(Equal("http://example.com/org1/repo1"))
	})

	It("Forgets deleted repos when a new location cache is generated", func() {
		fakeRepoService.ListByOrgStub = func(org string, _ *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error) {
			if org == "org1" {
				name := "first-repo"
				url := "http://example.com/org1/first-repo"
				return []*github.Repository{{Name: &name, HTMLURL: &url}}, &github.Response{}, nil
			}

			if org == "org2" {
				name := "repo1"
				url := "http://example.com/org2/repo1"
				return []*github.Repository{{Name: &name, HTMLURL: &url}}, &github.Response{}, nil
			}
			return nil, nil, errors.New("not found")
		}

		ifrit.Invoke(cacheLoader)
		_, firstfoundInCache := locCache.Lookup("first-repo")
		Expect(firstfoundInCache).To(BeTrue())

		fakeRepoService.ListByOrgStub = func(org string, _ *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error) {
			if org == "org1" {
				name := "second-repo"
				url := "http://example.com/org1/second-repo"
				return []*github.Repository{{Name: &name, HTMLURL: &url}}, &github.Response{}, nil
			}

			if org == "org2" {
				name := "repo1"
				url := "http://example.com/org2/repo1"
				return []*github.Repository{{Name: &name, HTMLURL: &url}}, &github.Response{}, nil
			}
			return nil, nil, errors.New("not found")
		}

		fakeClock.WaitForWatcherAndIncrement(cache.CacheUpdateInterval)
		Eventually(func() bool { _, found := locCache.Lookup("first-repo"); return found }).Should(BeFalse())
	})
})
