package config

import (
	"fmt"
	"strconv"
	"time"
)

// user=jack password=secret host=pg.example.com port=5432 dbname=mydb sslmode=verify-ca pool_max_conns=10
// postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca&pool_max_conns=10

// pool_max_conns             defaultMaxConns              int32(4)
// pool_min_conns             defaultMinConns              int32(0)
// pool_max_conn_lifetime     defaultMaxConnLifetime       time.Hour
// pool_max_conn_idle_time    defaultMaxConnIdleTime       time.Minute * 30
// pool_health_check_period   defaultHealthCheckPeriod     time.Minute

// n, err := strconv.ParseInt(s, 10, 32)
// config.MinConns = int32(n)
// config.MinConns = defaultMinConns

// d, err := time.ParseDuration(s)
// config.MaxConnIdleTime = defaultMaxConnIdleTime

type Postgre struct {
	Enable bool   `yaml:"enable"`
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`

	MaxConns uint32 `yaml:"pool_max_conns"` // int32(4)
	MinConns uint32 `yaml:"pool_min_conns"` // int32(0)

	PoolMaxConnLifetime string `yaml:"pool_max_conn_lifetime"`   // "1h"
	PoolMaxConnIdleTime string `yaml:"pool_max_conn_idle_time"`  // "30m"
	PoolHealthPeriod    string `yaml:"pool_health_check_period"` // "1m"

	Dsn string
}

func (cfg *Postgre) Options(db, user, password string) error {
	if len(cfg.Host) > 0 {
		cfg.Dsn += " host=" + cfg.Host
	}

	if len(cfg.Port) > 0 {
		cfg.Dsn += " port=" + cfg.Port
	}

	if len(user) > 0 {
		cfg.Dsn += " user=" + user
	}

	if len(password) > 0 {
		cfg.Dsn += " password=" + password
	}

	if len(db) > 0 {
		cfg.Dsn += " dbname=" + db
	}

	if cfg.MaxConns > 0 {
		cfg.Dsn += " pool_max_conns=" + strconv.FormatUint(uint64(cfg.MaxConns), 10)
	}

	if cfg.MinConns > 0 {
		cfg.Dsn += " pool_min_conns=" + strconv.FormatUint(uint64(cfg.MinConns), 10)
	}

	if len(cfg.PoolMaxConnLifetime) > 0 {
		t, err := time.ParseDuration(cfg.PoolMaxConnLifetime)
		if err != nil {
			return fmt.Errorf("invalid pool_max_conn_lifetime: %s", err.Error())
		}
		cfg.Dsn += " pool_max_conn_lifetime=" + t.String()
	}

	if len(cfg.PoolMaxConnIdleTime) > 0 {
		t, err := time.ParseDuration(cfg.PoolMaxConnIdleTime)
		if err != nil {
			return fmt.Errorf("invalid pool_max_conn_idle_time: %s", err.Error())
		}
		cfg.Dsn += " pool_max_conn_idle_time=" + t.String()
	}

	if len(cfg.PoolHealthPeriod) > 0 {
		t, err := time.ParseDuration(cfg.PoolHealthPeriod)
		if err != nil {
			return fmt.Errorf("invalid pool_health_check_period: %s", err.Error())
		}
		cfg.Dsn += " pool_health_check_period=" + t.String()
	}

	return nil
}
