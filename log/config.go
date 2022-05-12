package log

import (
	"io"
)

// --------------------------------
//    Config Log
// --------------------------------

type Config struct {
	// Разрешение на вывод отладочной информации.
	Debug bool
	Out   io.Writer
	// Формат ведения лога сообщений о ошибке. Возможные значения:
	// std  - вывод лога в поток io.Writer (значение по умолчанию os.Stderr),
	// net  - отправка сообщений на сервер RPC,
	// file - запись в текстовый файл.
	Format string
	// Путь к лог файлу.
	LogFile string
}

// Значения по умолчанию для работы с логом.
func NewDefaultConfig() *Config {
	return &Config{
		Debug:  false,
		Format: "std",
	}
}
