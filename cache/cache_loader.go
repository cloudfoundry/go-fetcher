package cache

import (
	"os"

	"github.com/google/go-github/github"
	"github.com/tedsuo/ifrit"
)

type cacheLoader struct {
	repoService RepositoriesService
	orgs        []string
}

//go:generate counterfeiter -o fakes/fake_repositories_service.go . RepositoriesService
type RepositoriesService interface {
	ListByOrg(org string, opt *github.RepositoryListByOrgOptions) ([]github.Repository, *github.Response, error)
}

func NewCacheLoader(orgs []string, repoService RepositoriesService) ifrit.Runner {
	return &cacheLoader{
		orgs:        orgs,
		repoService: repoService,
	}
}

func (c *cacheLoader) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	opt := &github.RepositoryListByOrgOptions{}
	for idx := range c.orgs {
		for {
			_, resp, err := c.repoService.ListByOrg(c.orgs[idx], opt)
			if err != nil {
				return err
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
