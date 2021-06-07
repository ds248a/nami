package log

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	gLogFormat   = map[string]uint8{"net": 1, "postgre": 1, "file": 1}
	errLogConDB  = errors.New("Error connect to Log DB")
	errLogFormat = errors.New("Error log format upload")
	errLogFile   = errors.New("Error log file")
	defFileMode  = os.FileMode(0644)
	defFileFlag  = os.O_RDWR | os.O_CREATE | os.O_APPEND
	sqlErrNoRows = "no rows in result set"
)

var lg *Logger

func init() {
	ctx, cancel := context.WithCancel(context.Background())

	lg = &Logger{
		Debug:    false,
		out:      os.Stderr,
		flag:     LstdFlags,
		format:   "file",
		fname:    "nami.log",
		Ctx:      ctx,
		Cancel:   cancel,
		ChMsg:    make(chan *Message, 1000),
		ChLen:    make(chan int),
		chBackup: make(chan string),
		Closer:   false,
	}

	if err := lg.open(); err != nil {
		Fatal(err)
	}

	// запуск сборщика логов
	go lg.save()
}

// --------------------------------
//    Log Init
// --------------------------------

func Format() string {
	return lg.format
}

// Регистрация настроек обработчика сообщений.
func NewLog(cfg *Config) error {
	if _, ok := gLogFormat[cfg.Format]; !ok {
		return errLogFormat
	}

	lg.mu.Lock()
	lg.Debug = cfg.Debug
	lg.format = cfg.Format

	// формат отправки соощений в базу данных Postgre
	if cfg.Format == "postgre" {
		if cfg.PDB == nil {
			return errLogConDB
		}
		lg.pdb = cfg.PDB
	}

	lg.mu.Unlock()

	// формат записи сообщений в текстовый файл
	if cfg.Format == "file" {
		if len(cfg.LogFile) > 0 {
			// переименование лог файла
			if err := lg.rename(cfg.LogFile); err != nil {
				return err
			}
		}

	} else {
		// предварительная обработка лог файла
		if err := lg.read(cfg); err != nil {
			return err
		}
	}

	return nil
}

// Обработка завершения работы приложения.
func Close(ct context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	Debug("-- Log Close")
	lg.Closer = true

	if len(lg.ChMsg) > 0 {
	loop:
		for {
			select {
			// ожидание таймера уровня приложения
			case <-ct.Done():
				break loop
			// ожидание освобождения канала
			case n := <-lg.ChLen:
				if n == 0 {
					break loop
				}
			}
		}
	}

	// nami.go - "test 1" + "test 3"
	// LogMsg("test 2").Save()

	lg.Cancel()
}

// --------------------------------
//    Log
// --------------------------------

// Формирование сообщения о ошибке.
func Err(err error) *Message {
	if err == nil {
		return nil
	}
	return newMessage(err.Error())
}

// Формирование сообщения на основе текстового сообщения.
func Msg(msg string) *Message {
	return newMessage(msg)
}

// Форматированный вывод отладочной информации.
func Debug(format string, args ...interface{}) {
	fmt.Printf(format+" \n", args...)
}

// Эквивалентна выполению logStd(), с последующим вызовом os.Exit(1).
func Fatal(err error) {
	lg.logOut(newMessage(err.Error()))
	os.Exit(1)
}

// Формирование сообщения в случае ошибки выполнения SQL запроса.
func DbErr(err error) *Message {
	if err != nil {
		if err.Error() == sqlErrNoRows {
			return nil
		}
		return newMessage(err.Error())
	}
	return nil
}

// В случае ошибки SQL запроса выполняется отмена транзакции,
// с последующим формированием сообщения о ошибке.
func DbTxErr(ctx context.Context, tx *pgxpool.Tx, err error) *Message {
	if err != nil {
		if err.Error() == sqlErrNoRows {
			return nil
		}
		DbTxRollback(ctx, tx)
		return newMessage(err.Error())
	}
	return nil
}

// Завершает текущую транзакцию или возвращает сообщение о ошибке.
func DbTxCommit(ctx context.Context, tx *pgxpool.Tx) *Message {
	if err := tx.Commit(ctx); err != nil {
		return newMessage(err.Error())
	}
	return nil
}

// Отменяет текущую транзакцию SQL запроса.
func DbTxRollback(ctx context.Context, tx *pgxpool.Tx) {
	if err := tx.Rollback(ctx); err != nil {
		if err.Error() != `tx is closed` {
			newMessage(err.Error()).Save()
		}
	}
}

// Определяет, привел ли SQL запрос к появлению ошибки.
func IsDbErr(err error) bool {
	if err != nil {
		if err.Error() == sqlErrNoRows {
			return false
		}
		return true
	}
	return false
}
