package log

import (
	//"bufio"
	"context"
	"encoding/json"
	"errors"

	//"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/ds248a/nami/config"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	gLogFormat   = map[string]uint8{"net": 1, "postgre": 1, "file": 1}
	errLogConDB  = errors.New("Error connect to Log DB")
	errLogFormat = errors.New("Error log format upload")
	defFileMode  = os.FileMode(0644)
	defFileFlag  = os.O_RDWR | os.O_CREATE | os.O_APPEND
)

var lg *Logger

// --------------------------------
//    Log Init
// --------------------------------

/*
type Logger struct {
	pdb   *pgxpool.Pool
	file  *os.File
	fname string

	Ctx      context.Context
	Cancel   context.CancelFunc
	ChData   chan *dbLog
	ChL      chan int
	chBackup chan string
	Closer   bool
}
*/

func NewLog(cfg *config.Loger, pdb *pgxpool.Pool) error {
	if _, ok := gLogFormat[cfg.Format]; !ok {
		return errLogFormat
	}

	if cfg.Format == "postgre" && pdb == nil {
		return errLogConDB
	}

	file, fname, err := logOpen(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	lg = &Logger{
		pdb:      pdb,
		file:     file,
		fname:    fname,
		Ctx:      ctx,
		Cancel:   cancel,
		ChData:   make(chan *dbLog, 1000),
		ChL:      make(chan int),
		chBackup: make(chan string),
		Closer:   false,
	}

	// предварительная обработка лог файла
	if err := lg.logRead(cfg); err != nil {
		return err
	}

	// запуск сборщика логов
	go logSave(cfg.Format)
	return nil
}

// --------------------------------

func logOpen(cfg *config.Loger) (*os.File, string, error) {
	fname := "./nami.log"
	if len(cfg.LogFile) > 0 {
		fname = cfg.LogFile
	}

	file, err := os.OpenFile(fname, defFileFlag, defFileMode)
	if err != nil {
		return nil, "", err
	}

	return file, fname, nil
}

// --------------------------------

func logSave(logFormat string) {
	defer func() { lg.ChData = nil }()

	for {
		select {
		case <-lg.Ctx.Done():
			return

		case d := <-lg.ChData:
			// режим остановки приложения: данные сохраняются в текстовый файл
			// данные из файла будут обработаны при следующем запуске сборщика логов
			if lg.Closer {
				logFile(d)
				lg.ChL <- len(lg.ChData)

			} else {
				// плановое сохранение данных, в соответствии с конфигурацией
				switch logFormat {
				case "net":
					logNet(d)
				case "postgre":
					logDb(d)
				case "file":
					logFile(d)
				default:
					logStd(d)
				}

			}
		}
	}
}

// --------------------------------

// обработка завершения работы приложения
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

func LogMsg(msg string) *dbLog {
	return logNew(msg)
}

// --------------------------------

func LogErr(err error) {
	logNew(err.Error())
}

// --------------------------------

func DbErr(err error) bool {
	if err != nil {
		// err != sql.ErrNoRows
		if err.Error() == `no rows in result set` {
			return true
		}
		logNew(err.Error())
		return true
	}
	return false
}

// --------------------------------

func DbTxErr(ctx context.Context, tx *pgxpool.Tx, err error) bool {
	if err != nil {
		// err != sql.ErrNoRows
		if err.Error() == `no rows in result set` {
			return true
		}
		DbTxRollback(ctx, tx)
		logNew(err.Error())
		return true
	}
	return false
}

// --------------------------------

func DbTxCommit(ctx context.Context, tx *pgxpool.Tx) bool {
	if err := tx.Commit(ctx); err != nil {
		logNew(err.Error())
		return false
	}
	return true
}

// --------------------------------

func DbTxRollback(ctx context.Context, tx *pgxpool.Tx) {
	if err := tx.Rollback(ctx); err != nil {
		if err.Error() != `tx is closed` {
			logNew(err.Error())
		}
	}
}

// --------------------------------

func IsDbErr(err error) bool {
	if err != nil {
		if err.Error() == `no rows in result set` {
			return false
		}
		return true
	}
	return false
}

// --------------------------------

func DbErrValid(err error) (bool, bool) {
	if err != nil {
		if err.Error() == `no rows in result set` {
			return false, false // совпадений не найдено && ошибок нет
		}
		logNew(err.Error())
		return false, true // ошибка в параметрах запроса
	}
	return true, false // возможно обнаружено совпадение && ошибок нет
}

// --------------------------------
//    Log Writer
// --------------------------------

// отправка записи сетевому сборщику
func logNet(l *dbLog) {
	// - ЗАГЛУШКА -
	// rpc.Dial()
	Debug("logNet [%s] line:%d file:%s \nerr:%s", l.Fnct, l.Line, l.File, l.Msg)
}

// --------------------------------

// регистрация записи в базе данных Postgre
func logDb(l *dbLog) {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	_, err := lg.pdb.Exec(ctx, `INSERT INTO main.log(file, line, function, message, datecreate) VALUES ($1, $2, $3, $4, $5)`, l.File, l.Line, l.Fnct, l.Msg, l.Date)
	if IsDbErr(err) {
		logFile(l)
	}
}

// --------------------------------

// регистрация записи в текстовом файле
func logFile(obj interface{}) {
	fp := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&lg.file)))
	file := (*os.File)(fp)
	if err := json.NewEncoder(file).Encode(&obj); err != nil {
		LogErr(err)
	}
}

// --------------------------------

// вывод сообщения в терминал
func logStd(l *dbLog) {
	Debug("logStd [%s] line:%d file:%s \nerr:%s", l.Fnct, l.Line, l.File, l.Msg)
}
