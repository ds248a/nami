package log

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"time"

	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	timeFormat = "2006/01/02 15:04:05"

	Ltime      = 1 << iota // time format "2006/01/02 15:04:05"
	Llongfile              // full file name and line number: /a/b/c/d.go:23
	Lshortfile             // final file name element and line number: d.go:23. overrides Llongfile
	LstdFlags  = Ltime | Lshortfile
)

// --------------------------------
//    Log
// --------------------------------

type Logger struct {
	mu sync.Mutex

	Debug bool      // вывод в терминал отладочной информации
	out   io.Writer // os.Strerr, по умолчанию
	flag  int       // LstdFlags, по умолчанию
	bufs  sync.Pool // буфер аварийного вывода. Используется при нарушении соединения определяемого параметром format

	format string        // формат оправки сообщений: rpc, postgre, file
	pdb    *pgxpool.Pool // соединение с базой данных Postgre
	file   *os.File      // структура лог файла
	fname  string        // путь лог файла

	Ctx      context.Context
	Cancel   context.CancelFunc
	ChMsg    chan *Message
	ChLen    chan int
	chBackup chan string
	Closer   bool
}

// Задает аварийный поток вывода ошибок.
// Используется при нарушении соединения определяемого параметром format.
// По умолчанию используется os.Stderr
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// FileInfo лог файла
func (l *Logger) Stat() (os.FileInfo, error) {
	return l.file.Stat()
}

// Открывает лог файл.
func (l *Logger) open() error {
	Debug("open")
	file, err := os.OpenFile(l.fname, defFileFlag, defFileMode)
	if err != nil {
		return err
	}

	l.file = file
	return nil
}

// Переименование лог файла
func (l *Logger) rename(fname string) error {
	Debug("rename")
	_, err := l.Stat()
	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	err = l.file.Close()
	if err != nil {
		return err
	}

	err = os.Rename(l.fname, fname)
	if err != nil {
		return err
	}

	l.fname = fname
	return l.open()
}

// Обработка аварийного лог файла.
func (l *Logger) read(cfg *Config) error {
	Debug("read")
	_, err := l.Stat()
	if err != nil {
		return err
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
func (l *Logger) save() {
	defer func() { l.ChMsg = nil }()

	for {
		select {
		case <-l.Ctx.Done():
			return

		case msg := <-l.ChMsg:
			// режим остановки приложения: данные сохраняются в текстовый файл
			// данные из файла будут обработаны при следующем запуске сборщика логов
			if l.Closer {
				l.logFile(msg)
				l.ChLen <- len(l.ChMsg)

			} else {
				Debug("logSave  format:%s  msg:%s", l.format, msg.Msg)
				// плановое сохранение данных, в соответствии с конфигурацией
				switch l.format {
				case "net":
					l.logNet(msg)
				case "postgre":
					l.logDb(msg)
				case "file":
					l.logFile(msg)
				default:
					l.logOut(msg)
				}
			}
		}
	}
}

// --------------------------------
//    Log Writer
// --------------------------------

// Отправка записи сетевому сборщику.
func (l *Logger) logNet(m *Message) {
	// rpc.Dial()
	Debug("logNet [%s] line:%d file:%s \nerr:%s", m.Fnct, m.Line, m.File, m.Msg)
}

// Отправка записи в базе данных Postgre.
func (l *Logger) logDb(m *Message) {
	Debug("logDb msg:%s", m.Msg)
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	_, err := l.pdb.Exec(ctx, `INSERT INTO main.log(file, line, function, message, datecreate) VALUES ($1, $2, $3, $4, $5)`, m.File, m.Line, m.Fnct, m.Msg, m.Date)
	if IsDbErr(err) {
		l.logFile(m)
	}
}

// Регистрация записи в текстовом файле.
func (l *Logger) logFile(m *Message) {
	Debug("logFile msg:%s", m.Msg)
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := json.NewEncoder(l.file).Encode(&m); err != nil {
		Fatal(err)
	}
}

// Отправка сообщения в поток вывода.
func (l *Logger) logOut(m *Message) error {
	Debug("logOut msg:%s", m.Msg)

	buf := l.bufs.Get().([]byte)
	buf = buf[0:0]
	defer l.bufs.Put(buf)

	if l.flag&Ltime > 0 {
		now := time.Now().Format(timeFormat)
		buf = append(buf, '[')
		buf = append(buf, now...)
		buf = append(buf, "] "...)
	}

	if l.flag&(Lshortfile|Llongfile) != 0 {
		buf = append(buf, m.File...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(m.Line), 10)
		buf = append(buf, ':')
		buf = append(buf, m.Fnct...)
		buf = append(buf, ' ')
	}

	buf = append(buf, m.Msg...)
	if len(m.Msg) == 0 || m.Msg[len(m.Msg)-1] != '\n' {
		buf = append(buf, '\n')
	}

	l.mu.Lock()
	_, err := l.out.Write(buf)
	l.mu.Unlock()

	return err
}
