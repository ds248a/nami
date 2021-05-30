package log

import (
	//"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	//"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	// "github.com/ds248a/nami/config"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	gLogFormat   = map[string]uint8{"net": 1, "postgre": 1, "file": 1}
	errLogConDB  = errors.New("Error connect to Log DB")
	errLogFormat = errors.New("Error log format upload")
	defFileMode  = os.FileMode(0644)
	defFileFlag  = os.O_RDWR | os.O_CREATE | os.O_APPEND
	sqlErrNoRows = "no rows in result set"
)

var lg *Logger

func init() {
	ctx, cancel := context.WithCancel(context.Background())

	lg = &Logger{
		Debug:    false,
		format:   "file",
		fname:    "./nami.log",
		Ctx:      ctx,
		Cancel:   cancel,
		ChData:   make(chan *Message, 1000),
		ChL:      make(chan int),
		chBackup: make(chan string),
		Closer:   false,
	}

	if err := lg.logOpen(); err != nil {
		Fatal(err)
	}

	// запуск сборщика логов
	go lg.logSave()
}

// --------------------------------
//    Log Init
// --------------------------------

func Format() string {
	return lg.format
}

// Регистрация настроек обработчика сообщений.
// func NewLog(cfg *config.Logger, pdb *pgxpool.Pool) error {
func NewLog(cfg *Config) error {
	if _, ok := gLogFormat[cfg.Format]; !ok {
		return errLogFormat
	}

	lg.Debug = cfg.Debug
	lg.format = cfg.Format

	// формат отправки соощений в базу данных Postgre
	if cfg.Format == "postgre" {
		if cfg.PDB == nil {
			return errLogConDB
		}
		lg.pdb = cfg.PDB
	}

	// предварительная обработка лог файла
	if err := lg.logRead(cfg); err != nil {
		return err
	}

	return nil
}

// Обработка завершения работы приложения.
func LogClose(ct context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	Debug("-- Log Close")
	lg.Closer = true

	if len(lg.ChData) > 0 {
	loop:
		for {
			select {
			// ожидание таймера уровня приложения
			case <-ct.Done():
				break loop
			// ожидание освобождения канала
			case n := <-lg.ChL:
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

// Форматированный вывод отладочной информации.
func Debug(format string, args ...interface{}) {
	fmt.Printf(format+" \n", args...)
}

// Эквивалентна выплению logStd(), с последующим вызовом os.Exit(1).
func Fatal(err error) {
	logStd(newMessage(err.Error()))
	os.Exit(1)
}

// Формирование сообщения о ошибке.
func Msg(msg string) *Message {
	return newMessage(msg)
}

// Формирование сообщения о ошибке.
func Err(err error) *Message {
	return newMessage(err.Error())
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

// --------------------------------
//    Log Writer
// --------------------------------

// Отправка записи сетевому сборщику.
func logNet(l *Message) {
	// rpc.Dial()
	Debug("logNet [%s] line:%d file:%s \nerr:%s", l.Fnct, l.Line, l.File, l.Msg)
}

// Регистрация записи в базе данных Postgre.
func logDb(l *Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	_, err := lg.pdb.Exec(ctx, `INSERT INTO main.log(file, line, function, message, datecreate) VALUES ($1, $2, $3, $4, $5)`, l.File, l.Line, l.Fnct, l.Msg, l.Date)
	if IsDbErr(err) {
		logFile(l)
	}
}

// Регистрация записи в текстовом файле.
func logFile(obj interface{}) {
	fp := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&lg.file)))
	file := (*os.File)(fp)
	if err := json.NewEncoder(file).Encode(&obj); err != nil {
		Fatal(err)
	}
}

// Вывод сообщения в терминал.
func logStd(l *Message) {
	Debug("logStd [%s] line:%d file:%s \nerr:%s", l.Fnct, l.Line, l.File, l.Msg)
}
