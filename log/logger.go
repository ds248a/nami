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
	ChData   chan *Message
	ChL      chan int
	chBackup chan string
	Closer   bool
}

// Открывает лог файл.
func (l *Logger) logOpen() error {
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

	fback := l.fname + ".backup"
	if err := os.Rename(l.fname, fback); err != nil {
		return err
	}

	if len(cfg.LogFile) > 0 {
		l.fname = cfg.LogFile
	}

	// регистрация нового лог файла
	newFile, err := os.OpenFile(l.fname, defFileFlag, defFileMode)
	if err != nil {
		return err
	}

	oldFile := atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&l.file)), unsafe.Pointer(newFile))

	// выгрузка лог файла
	go func() {
		defer func() {
			if err := (*os.File)(oldFile).Close(); err != nil {
				Err(err).Save()
			}
			if err := os.Remove(fback); err != nil {
				Err(err).Save()
			}
		}()

		r := bufio.NewReader((*os.File)(oldFile))

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
	defer func() { l.ChData = nil }()

	for {
		select {
		case <-l.Ctx.Done():
			return

		case d := <-l.ChData:
			// режим остановки приложения: данные сохраняются в текстовый файл
			// данные из файла будут обработаны при следующем запуске сборщика логов
			if l.Closer {
				logFile(d)
				l.ChL <- len(l.ChData)

			} else {
				// плановое сохранение данных, в соответствии с конфигурацией
				switch l.format {
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
