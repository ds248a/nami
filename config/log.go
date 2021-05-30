package config

type Logger struct {
	Format  string `yaml:"format"`
	LogFile string `yaml:"log_file"`
}
