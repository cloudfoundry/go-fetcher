package cache_test

import (
	"github.com/cloudfoundry/go-fetcher/cache"
	"github.com/cloudfoundry/go-fetcher/cache/fakes"
	"github.com/google/go-github/github"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CacheLoader", func() {
	var (
		fakeRepoService *fakes.FakeRepositoriesService
		cacheLoader     ifrit.Runner
	)

	BeforeEach(func() {
		fakeRepoService = &fakes.FakeRepositoriesService{}
		cacheLoader = cache.NewCacheLoader([]string{"org1", "org2"}, fakeRepoService)
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
})
