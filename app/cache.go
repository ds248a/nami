package app

import (
	"context"
	"errors"
	"sync"

	"github.com/ds248a/air/config"
	"github.com/patrickmn/go-cache"
)

var (
	errCacheConfig = errors.New("Local Cache not configured")
)

// --------------------------------
//    Local Cache
// --------------------------------

func newCache(cfg *config.Cache) error {
	if cfg == nil {
		return errCacheConfig
	}

	if err := cfg.Options(); err != nil {
		return err
	}
	lc = cache.New(cfg.Expire, cfg.Clean)

	callOnExit(func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		dst := make(chan bool)
		go func(lc *cache.Cache) {
			lc.Flush()
			Debug("-- Cache Close")
			dst <- true
		}(lc)

	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-dst:
				break loop
			}
		}
	})

	return nil
}

func Cache() *cache.Cache {
	return lc
}
