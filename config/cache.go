package config

import (
	"fmt"
	"time"
)

// Параметры времени указывать в секундах.

const (
	cacheExpire = time.Second * 600
	cacheClean  = time.Second * 900
)

type Cache struct {
	ExpTimeout   string `yaml:"expiration"`
	CleanTimeout string `yaml:"cleanup"`
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
