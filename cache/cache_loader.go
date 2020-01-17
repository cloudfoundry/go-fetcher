package cache

import (
	"context"
	"os"
	"time"

	"github.com/google/go-github/github"
	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"github.com/tedsuo/ifrit"
)

const CacheUpdateInterval = 10 * time.Minute

type cacheLoader struct {
	logger        lager.Logger
	githubURL     string
	orgs          []string
	locationCache *LocationCache
	repoService   RepositoriesService
	clock         clock.Clock
}

//go:generate counterfeiter -o fakes/fake_repositories_service.go . RepositoriesService
type RepositoriesService interface {
	ListByOrg(ctx context.Context, org string, opt *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error)
}

func NewCacheLoader(logger lager.Logger, githubURL string, orgs []string, locationCache *LocationCache, repoService RepositoriesService, clock clock.Clock) ifrit.Runner {
	return &cacheLoader{
		logger:        logger,
		githubURL:     githubURL,
		orgs:          orgs,
		locationCache: locationCache,
		repoService:   repoService,
		clock:         clock,
	}
}

func (c *cacheLoader) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := c.logger

	// Initialize the cache
	err := c.updateCache(logger)

	// On starup, fail if there is an error with the initial call to github,
	// becaue it's more likely to be noticed and there's a higher change the
	// problem is with us. During the regular interval updates of the change,
	// don't bring the process down if there's an error talking to github, as we
	// expect it might just be temporary downtime for github.
	if err != nil {
		logger.Error("failed-starting-cache-loader", err)
		return err
	}

	close(ready)

	timer := c.clock.NewTimer(CacheUpdateInterval)
	for {
		select {
		case <-timer.C():
			err := c.updateCache(logger)
			if err != nil {
				logger.Error("failed-updating-cache", err)
			}
			timer.Reset(CacheUpdateInterval)
		case signal := <-signals:
			logger.Info("signaled", lager.Data{"signal": signal.String})
			timer.Stop()
		}
	}
	return nil
}

func (c *cacheLoader) updateCache(logger lager.Logger) error {
	logger = logger.Session("update-cache")
	logger.Info("fetching-orgs", lager.Data{"orgs": c.orgs})

	tempLocationCache := NewLocationCache(c.logger, c.clock)
	for i := len(c.orgs) - 1; i >= 0; i-- {
		org := c.orgs[i]
		logger.Info("fetching-org", lager.Data{"org": org})
		opt := &github.RepositoryListByOrgOptions{
			Type:        "public",
			ListOptions: github.ListOptions{PerPage: 100, Page: 1},
		}

		for {
			logger.Info("fetching-page", lager.Data{"org": org, "page": opt.Page})
			repos, resp, err := c.repoService.ListByOrg(context.Background(), org, opt)
			if err != nil {
				logger.Error("failed-fetching-page", err, lager.Data{"org": org, "page": opt.Page})
				return err
			}

			for _, repo := range repos {
				logger.Debug("found-repo", lager.Data{"repo": repo.Name})
				if repo.Name == nil {
					continue
				}

				tempLocationCache.Add(*repo.Name, *repo.HTMLURL)
			}

			logger.Info("finished-page", lager.Data{"org": org, "page": opt.Page, "next": resp.NextPage, "last": resp.LastPage})
			if resp.NextPage == 0 {
				break
			}
			opt.ListOptions.Page = resp.NextPage
		}
	}
	logger.Info("finished-fetching-orgs", lager.Data{"orgs": c.orgs})

	c.locationCache.Swap(tempLocationCache)

	return nil
}
