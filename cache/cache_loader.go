package cache

import (
	"os"

	"github.com/tedsuo/ifrit"
)

type cacheLoader struct {
}

func NewCacheLoader() ifrit.Runner {
	return &cacheLoader{}
}

func (c *cacheLoader) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	close(ready)

	<-signals
	return nil
}
