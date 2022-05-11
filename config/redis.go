package config

import (
	"time"

	"github.com/go-redis/redis/v8"
)

// -------------------------------------
// параметры времени задавать в секундах
// -------------------------------------

type RedisRing struct {
	Enable       bool              `yaml:"enable"`
	Addrs        map[string]string `yaml:"addrs"`
	Password     string            `yaml:"password"`
	DB           int               `yaml:"db"`
	PoolSize     int               `yaml:"pool_size"`
	MaxRetries   int               `yaml:"max_retries"`
	PoolTimeout  string            `yaml:"pool_timeout"`
	IdleTimeout  string            `yaml:"idle_timeout"`
	DialTimeout  string            `yaml:"dial_timeout"`
	ReadTimeout  string            `yaml:"read_timeout"`
	WriteTimeout string            `yaml:"write_timeout"`
}

func (cfg *RedisRing) Options() *redis.RingOptions {
	opt := &redis.RingOptions{
		Addrs:    cfg.Addrs,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	}

	if cfg.MaxRetries > 0 {
		opt.MaxRetries = cfg.MaxRetries
	}

	if len(cfg.PoolTimeout) > 0 {
		if t, err := time.ParseDuration(cfg.PoolTimeout); err == nil {
			opt.PoolTimeout = t
		}
	}

	if len(cfg.IdleTimeout) > 0 {
		if t, err := time.ParseDuration(cfg.IdleTimeout); err == nil {
			opt.IdleTimeout = t
		}
	}

	if len(cfg.DialTimeout) > 0 {
		if t, err := time.ParseDuration(cfg.DialTimeout); err == nil {
			opt.DialTimeout = t
		}
	}

	if len(cfg.ReadTimeout) > 0 {
		if t, err := time.ParseDuration(cfg.ReadTimeout); err == nil {
			opt.ReadTimeout = t
		}
	}

	if len(cfg.WriteTimeout) > 0 {
		if t, err := time.ParseDuration(cfg.WriteTimeout); err == nil {
			opt.WriteTimeout = t
		}
	}

	return opt
}
