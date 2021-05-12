package config

import (
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var (
	envFlag    = flag.String("env", "dev", "environment: dev or prod")
	appDirFlag = flag.String("app_dir", "config", "config dir for *.yml file")
	errSrvAdr  = errors.New("HTTP Server address not valid")
)

type Config struct {
	AppDir string
	Env    string

	ServerAdr string     `yaml:"server_addr"`
	Postgre   *Postgre   `yaml:"postgre"`
	Redis     *RedisRing `yaml:"redis"`
	Cache     *Cache     `yaml:"cache"`

	Debug bool   `yaml:"debug"`
	Loger *Loger `yaml:"log"`
}

func init() {
	flag.Parse()
}

func LoadConfig() (*Config, error) {
	return loadConfigEnv(*appDirFlag, *envFlag)
}

func LoadConfigEnv(env string) (*Config, error) {
	return loadConfigEnv(*appDirFlag, env)
}

func loadConfigEnv(appDir, env string) (*Config, error) {
	appDir, err := filepath.Abs(appDir)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filepath.Join(appDir, env+".yml"))
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	cfg, err := parseConfig(b)
	if err != nil {
		return nil, err
	}

	if len(cfg.ServerAdr) == 0 {
		return nil, errSrvAdr
	}

	cfg.AppDir = appDir
	cfg.Env = env

	return cfg, nil
}

func parseConfig(b []byte) (*Config, error) {
	cfg := new(Config)
	err := yaml.Unmarshal(b, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

/*
debug: true
server_addr: "172.16.10.211:8010"
redis:
  addrs:
    server1: "172.16.10.211:6379"
  db: 0
  pool_size: 8
postgre:
  host: "172.16.10.211"
  port: "5432"
  pool_max_conns: 5
cache:
  expiration: "10m"
  cleanup: "15m"
log:
  format: "postgre"
  table: "log"
*/
