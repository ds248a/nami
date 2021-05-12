package app

import (
	"context"
	"errors"
	"sync"

	"github.com/ds248a/nami/config"
	lg "github.com/ds248a/nami/log"
	"github.com/go-redis/redis/v8"
)

var (
	errRedisConfig = errors.New("Redis not configured")
)

// --------------------------------
//    Redis
// --------------------------------

func newRedis(cfg *config.RedisRing) error {
	if cfg == nil {
		return errRedisConfig
	}

	rdb = redis.NewRing(cfg.Options())
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return err
	}

	callOnExit(func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		dst := make(chan bool)
		go func(db *redis.Ring) {
			if err := db.Close(); err != nil {
				lg.LogErr(err)
			}
			dst <- true
			Debug("-- Redis Close")
		}(rdb)

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

func Redis() *redis.Ring {
	return rdb
}
