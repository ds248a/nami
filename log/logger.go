package log

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"

	// "sync"
	"sync/atomic"
	"unsafe"

	// "github.com/ds248a/nami/config"
	"github.com/jackc/pgx/v4/pgxpool"
)

// --------------------------------
//    Log
// --------------------------------

type Logger struct {
	// mu sync.Mutex
	Debug bool

	format string
	pdb    *pgxpool.Pool
	file   *os.File
	fname  string

	Ctx      context.Context
	Cancel   context.CancelFunc
	ChMsg    chan *Message
	ChLen    chan int
	chBackup chan string
	Closer   bool
}

// Открывает лог файл.
func (l *Logger) logOpen() error {
	Debug("logOpen")
	file, err := os.OpenFile(l.fname, defFileFlag, defFileMode)
	if err != nil {
		return err
	}

	l.file = file
	return nil
}

// Обработка аварийного лог файла.
func (l *Logger) logRead(cfg *Config) error {
	if cfg.Format == "file" {
		return nil
	}

	fInfo, err := l.file.Stat()
	if err != nil {
		return err
	}
	if fInfo.Size() == 0 {
		return nil
	}

	// создание временного файла лога
	fbackup := l.fname + ".backup"
	if err := os.Rename(l.fname, fbackup); err != nil {
		return err
	}

	// если конфигурация переопределяет файл лога
	if len(cfg.LogFile) > 0 {
		l.fname = cfg.LogFile
	}

	// регистрация нового лог файла
	newFile, err := os.OpenFile(l.fname, defFileFlag, defFileMode)
	if err != nil {
		return err
	}

	// атомарная подмена файла
	oldFile := (*os.File)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&l.file)), unsafe.Pointer(newFile)))

	// выгрузка лог файла
	go func() {
		defer func() {
			if err := oldFile.Close(); err != nil {
				Err(err).Save()
			}
			if err := os.Remove(fbackup); err != nil {
				Err(err).Save()
			}
		}()

		if _, err = oldFile.Seek(0, 0); err != nil {
			Err(err).Save()
			return
		}

		r := bufio.NewReader(oldFile)

		for {
			line, err := r.ReadString('\n')

			if err == io.EOF {
				break
			} else if err != nil {
				Err(err).Save()
				return
			}

			msg := &Message{}
			if err := json.Unmarshal([]byte(line), msg); err != nil {
				Err(err).Save()
				return
			}

			// регистрация записи в соответствии с настройками конфигурации
			msg.Save()
		}
	}()

	return nil
}

// Запрос на сохранение сообщения в соответствии с конфигурацией.
func (l *Logger) logSave() {
	defer func() { l.ChMsg = nil }()

	for {
		select {
		case <-l.Ctx.Done():
			return

		case msg := <-l.ChMsg:
			// режим остановки приложения: данные сохраняются в текстовый файл
			// данные из файла будут обработаны при следующем запуске сборщика логов
			if l.Closer {
				msg.logFile()
				l.ChLen <- len(l.ChMsg)

			} else {
				Debug("logSave  format:%s  msg:%s", l.format, msg.Msg)
				// плановое сохранение данных, в соответствии с конфигурацией
				switch l.format {
				case "net":
					msg.logNet()
				case "postgre":
					msg.logDb()
				case "file":
					msg.logFile()
				default:
					msg.logStd()
				}
			}
		}
	}
}
