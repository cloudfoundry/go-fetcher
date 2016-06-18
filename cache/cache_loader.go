package cache

import (
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.com/pivotal-golang/clock"
	"github.com/tedsuo/ifrit"
)

const CacheUpdateInterval = 10 * time.Minute

type cacheLoader struct {
	repoService   RepositoriesService
	orgs          []string
	locationCache *LocationCache
	clock         clock.Clock
}

//go:generate counterfeiter -o fakes/fake_repositories_service.go . RepositoriesService
type RepositoriesService interface {
	ListByOrg(org string, opt *github.RepositoryListByOrgOptions) ([]github.Repository, *github.Response, error)
}

func NewCacheLoader(orgs []string, locationCache *LocationCache, repoService RepositoriesService, clock clock.Clock) ifrit.Runner {
	return &cacheLoader{
		orgs:          orgs,
		locationCache: locationCache,
		repoService:   repoService,
		clock:         clock,
	}
}

func (c *cacheLoader) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	err := c.updateCache()

	// On starup, fail if there is an error with the initial call to github,
	// becaue it's more likely to be noticed and there's a higher change the
	// problem is with us. During the regular interval updates of the change, don't
	// bring the process down if there's an error talking to github, as we expect
	// it might just be temporary downtime for github.
	if err != nil {
		return err
	}

	close(ready)

	timer := c.clock.NewTimer(CacheUpdateInterval)
	for {
		select {
		case <-timer.C():
			_ = c.updateCache()
			// TODO: log error if non-nil
			timer.Reset(CacheUpdateInterval)

		case <-signals:
			// TODO: log that we were signalled
		}
	}
	return nil
}

func (c *cacheLoader) updateCache() error {
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

	return nil
}
