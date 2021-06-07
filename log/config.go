package log

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

// --------------------------------
//    Config Log
// --------------------------------

type Config struct {
	// Разрешение на вывод отладочной информации.
	Debug bool
	// Формат ведения лога сообщений о ошибке. Возможные значения:
	// net     - отправка на сервер RPC;
	// postgre - запись в таблицу СУБД Postgre;
	// file    - запись в текстовый файл.
	Format string
	// Соединение с БД Postgre.
	PDB *pgxpool.Pool
	// Путь к лог файлу.
	// 'nami.log' - значение по умодчанию.
	LogFile string
}

// По умолчанию сообщения сохряняются в текстовом файле 'nami.log'.
// Ротация лог файла не предусмотрена.
func NewDefaultConfig() *Config {
	return &Config{
		Debug:   false,
		Format:  "file",
		LogFile: "nami.log",
	}
}
