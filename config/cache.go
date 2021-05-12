package config

import (
	"fmt"
	"time"
)

// time.Minute

var (
	cacheExpire = time.Second * 600
	cacheClean  = time.Second * 900
)

// -------------------------------------
// параметры времени указывать в секундах
// -------------------------------------

type Cache struct {
	ExpTimeout   string `yaml:"expiration"` // "10m"
	CleanTimeout string `yaml:"cleanup"`    // "15m"
	Expire       time.Duration
	Clean        time.Duration
}

func (cfg *Cache) Options() error {
	if len(cfg.ExpTimeout) > 0 {
		t, err := time.ParseDuration(cfg.ExpTimeout)
		if err != nil {
			return fmt.Errorf("Cache expiration: %s", err.Error())
		}
		cfg.Expire = t
	} else {
		cfg.Expire = cacheExpire
	}

	if len(cfg.CleanTimeout) > 0 {
		t, err := time.ParseDuration(cfg.CleanTimeout)
		if err != nil {
			return fmt.Errorf("Cache cleanup: %s", err.Error())
		}
		cfg.Clean = t
	} else {
		cfg.Clean = cacheClean
	}

	return nil
}
