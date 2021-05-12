package log

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync/atomic"
	"unsafe"

	"github.com/ds248a/nami/config"
	"github.com/jackc/pgx/v4/pgxpool"
)

// --------------------------------
//    Logger
// --------------------------------

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

// обработка аварийного лог файла
func (l *Logger) logRead(cfg *config.Loger) error {
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

	// регистрация нового лог файла
	newFile, err := os.OpenFile(l.fname, defFileFlag, defFileMode)
	if err != nil {
		return err
	}

	oldFile := atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&l.file)), unsafe.Pointer(newFile))

	// выгрузка лог файла
	go func() {
		defer (*os.File)(oldFile).Close()
		r := bufio.NewReader((*os.File)(oldFile))

		for {
			line, err := r.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				LogErr(err)
				return
			}

			jl := &dbLog{}
			if err := json.Unmarshal([]byte(line), jl); err != nil {
				LogErr(err)
				return
			}

			// регистрация записи в соответствии с настройками конфигурации
			jl.Save()
		}

		if err := os.Remove(fback); err != nil {
			LogErr(err)
		}
	}()

	return nil
}
