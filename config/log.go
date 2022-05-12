package config

// Струкура определяет формат ведения лога. Возможные значения:
// std     - вывод лога в стандартный поток os.Stderr (значение по умолчанию),
// net     - отправка сообщений на сервер RPC,
// file    - запись в текстовый файл.
type Logger struct {
	Format  string `yaml:"format"`
	LogFile string `yaml:"log_file"`
}
