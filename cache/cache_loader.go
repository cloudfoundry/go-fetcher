package cache

import (
	"os"

	"github.com/google/go-github/github"
	"github.com/tedsuo/ifrit"
)

type cacheLoader struct {
	repoService   RepositoriesService
	orgs          []string
	locationCache *LocationCache
}

//go:generate counterfeiter -o fakes/fake_repositories_service.go . RepositoriesService
type RepositoriesService interface {
	ListByOrg(org string, opt *github.RepositoryListByOrgOptions) ([]github.Repository, *github.Response, error)
}

func NewCacheLoader(orgs []string, locationCache *LocationCache, repoService RepositoriesService) ifrit.Runner {
	return &cacheLoader{
		orgs:          orgs,
		locationCache: locationCache,
		repoService:   repoService,
	}
}

func (c *cacheLoader) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	opt := &github.RepositoryListByOrgOptions{}
	for _, org := range c.orgs {
		for {
			repos, resp, err := c.repoService.ListByOrg(org, opt)
			if err != nil {
				return err
			}
			for _, repo := range repos {
				if repo.Name == nil {
					continue
				}

				name := *(repo.Name)

				c.locationCache.Add(name, org+name)
			}
			if resp.NextPage == 0 {
				break
			}
			opt.ListOptions.Page = resp.NextPage
		}
	}

	close(ready)

	<-signals
	return nil
}
